// Copyright Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package com.google.samples.apps.gcmplayground;

import android.app.Activity;
import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.DialogInterface;
import android.content.Intent;
import android.content.IntentFilter;
import android.net.Uri;
import android.os.AsyncTask;
import android.os.Bundle;
import android.support.v4.content.LocalBroadcastManager;
import android.util.Log;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.TextView;
import android.widget.Toast;

import com.google.android.gms.common.ConnectionResult;
import com.google.android.gms.common.GooglePlayServicesUtil;
import com.google.samples.apps.gcmplayground.constants.RegistrationConstants;
import com.google.samples.apps.gcmplayground.util.GCMPlaygroundUtil;

import java.io.IOException;

public class MyActivity extends Activity  {

    private static final int PLAY_SERVICES_RESOLUTION_REQUEST = 9000;
    private static final String TAG = "MyActivity";

    private BroadcastReceiver mRegistrationBroadcastReceiver;
    private BroadcastReceiver mDownstreamBroadcastReceiver;
    private Button registerButton;
    private Button unregisterButton;
    private EditText senderIdField;
    private EditText appServerHostField;
    private EditText stringIdentifierField;
    private TextView registrationTokenFieldView;
    private TextView statusView;
    private TextView downstreamBundleView;
    private String token;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_my);

        registerButton = (Button) findViewById(R.id.register_button);
        unregisterButton = (Button) findViewById(R.id.unregister_button);
        senderIdField = (EditText) findViewById(R.id.sender_id);
        appServerHostField = (EditText) findViewById(R.id.app_server_host);
        stringIdentifierField = (EditText) findViewById(R.id.string_identifier);
        registrationTokenFieldView = (TextView) findViewById(R.id.registeration_token);
        statusView = (TextView) findViewById(R.id.status);
        downstreamBundleView = (TextView) findViewById(R.id.downstream_bundle);

        // If Play Services is not up to date, quit the app.
        checkPlayServices();

        // Restore from saved instance state
        if (savedInstanceState != null) {
            token = savedInstanceState.getString(RegistrationConstants.EXTRA_KEY_TOKEN, "");
            if (token != "") {
                updateUI("Registration SUCCEEDED", true);
            }
        }

        mRegistrationBroadcastReceiver = new BroadcastReceiver() {
            @Override
            public void onReceive(Context context, Intent intent) {
                boolean sentToken = intent.getBooleanExtra(
                        RegistrationConstants.SENT_TOKEN_TO_SERVER, false);

                if (sentToken) {
                    token = intent.getStringExtra(RegistrationConstants.EXTRA_KEY_TOKEN);
                    updateUI("Registration SUCCEEDED", true);
                } else {
                    updateUI("Registration FAILED", false);
                }
            }
        };

        mDownstreamBroadcastReceiver = new BroadcastReceiver() {
            @Override
            public void onReceive(Context context, Intent intent) {
                String from = intent.getStringExtra(RegistrationConstants.SENDER_ID);
                Bundle data = intent.getBundleExtra(RegistrationConstants.EXTRA_KEY_BUNDLE);
                String message = data.getString("message");

                downstreamBundleView.setText(data.toString());

                Log.d(TAG, "Received from >" + from + "< with >" + data.toString() + "<");
                Log.d(TAG, "Message: " + message);
            }
        };

        LocalBroadcastManager.getInstance(this).registerReceiver(mRegistrationBroadcastReceiver,
                new IntentFilter(RegistrationConstants.REGISTRATION_COMPLETE));
        LocalBroadcastManager.getInstance(this).registerReceiver(mDownstreamBroadcastReceiver,
                new IntentFilter(RegistrationConstants.NEW_DOWNSTREAM_MESSAGE));

        // TODO(karangoel): Remove these. Only for development purposes
        senderIdField.setText("436520785863");
        appServerHostField.setText("751cebd0.ngrok.io");
        stringIdentifierField.setText("Nexus 5");
    }

    @Override
    protected void onSaveInstanceState(Bundle outState) {
        super.onSaveInstanceState(outState);
        outState.putString(RegistrationConstants.EXTRA_KEY_TOKEN, token);
    }

    private void updateUI(String status, boolean registered) {
        // Set status and token text
        statusView.setText(status);
        registrationTokenFieldView.setText(token);

        // Button enabling
        registerButton.setEnabled(!registered);
        unregisterButton.setEnabled(registered);
    }

    @Override
    protected void onResume() {
        super.onResume();
        LocalBroadcastManager.getInstance(this).registerReceiver(mRegistrationBroadcastReceiver,
                new IntentFilter(RegistrationConstants.REGISTRATION_COMPLETE));
    }

    @Override
    protected void onPause() {
        LocalBroadcastManager.getInstance(this).unregisterReceiver(mRegistrationBroadcastReceiver);
        super.onPause();
    }

    @Override
    protected void onStop() {
        LocalBroadcastManager.getInstance(this).unregisterReceiver(mRegistrationBroadcastReceiver);
        super.onStop();
    }

    /**
     * Calls the GCM API to register this client if not already registered.
     * @param view
     * @throws IOException
     */
    public void registerClient(View view) throws IOException {
        // Get the sender ID
        String senderId = senderIdField.getText().toString();
        String stringId = stringIdentifierField.getText().toString();
        String host = appServerHostField.getText().toString();

        if (senderId == "" || host == "") {
            showToast("Sender ID and host cannot be empty.");
            return;
        }
        Log.d(TAG, senderId);
        // Register with GCM
        Intent intent = new Intent(this, RegistrationIntentService.class);
        intent.putExtra(RegistrationConstants.SENDER_ID, senderId);
        intent.putExtra(RegistrationConstants.STRING_IDENTIFIER, stringId);
        intent.putExtra(RegistrationConstants.HOST, host);
                startService(intent);
    }

    /**
     * Calls the GCM API to unregister this client
     * @param view
     */
    public void unregisterClient(View view) {
        String host = appServerHostField.getText().toString();
        Uri.Builder builder = new Uri.Builder();
        builder.scheme("http")
                .authority(host)
                .appendPath("clients")
                .appendPath(token);
        (new UnregisterClientTask()).execute(builder.build().toString());
    }

    private class UnregisterClientTask extends AsyncTask<String, Void, Integer> {

        @Override
        protected Integer doInBackground(String... urls) {
            return GCMPlaygroundUtil.delete(urls[0]);
        }

        @Override
        protected void onPostExecute(Integer code) {
            Log.d(TAG, Integer.toString(code));
            if (code != RegistrationConstants.VALID_DELETE_RESPONSE) {
                statusView.setText("Unregistration failed: " + code);
            } else {
                token = "";
                updateUI("Unregistration SUCCEEDED", false);
                showToast("Unregistered!");
            }
        }
    }

    /**
     * Show a toast with passed text
     * @param text to be used as toast message
     */
    public void showToast(CharSequence text) {
        Toast.makeText(this, text, Toast.LENGTH_SHORT).show();
    }

    /**
     * Check the device to make sure it has the Google Play Services APK. If
     * it doesn't, display a dialog that allows users to download the APK from
     * the Google Play Store or enable it in the device's system settings.
     */
    private void checkPlayServices() {
        int resultCode = GooglePlayServicesUtil.isGooglePlayServicesAvailable(this);
        if (resultCode != ConnectionResult.SUCCESS) {
            if (GooglePlayServicesUtil.isUserRecoverableError(resultCode)) {
                GooglePlayServicesUtil.getErrorDialog(resultCode, this,
                        PLAY_SERVICES_RESOLUTION_REQUEST,
                        new DialogInterface.OnCancelListener() {
                            @Override
                            public void onCancel(DialogInterface dialog) {
                                finish();
                            }
                        }).show();
            } else {
                Log.w(TAG, "Google Play Services is required and not supported on this device.");
            }
        }
    }

}
