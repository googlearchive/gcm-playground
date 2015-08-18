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

import android.app.IntentService;
import android.content.Intent;
import android.net.Uri;
import android.os.Bundle;
import android.support.v4.content.LocalBroadcastManager;
import android.util.Log;

import com.google.android.gms.gcm.GoogleCloudMessaging;
import com.google.android.gms.iid.InstanceID;
import com.google.samples.apps.gcmplayground.constants.RegistrationConstants;
import com.google.samples.apps.gcmplayground.util.GCMPlaygroundUtil;

import java.io.IOException;

public class RegistrationIntentService extends IntentService {

    private static final String TAG = "RegIntentService";

    public RegistrationIntentService() {
        super(TAG);
    }

    @Override
    protected void onHandleIntent(Intent intent) {
        Bundle extras = intent.getExtras();
        String token = "";

        Intent regCompleteIntent = new Intent(RegistrationConstants.REGISTRATION_COMPLETE);

        try {
            // Initially this call goes out to the network to retrieve the token, subsequent
            // calls are local.
            String sender_id = extras.getString(RegistrationConstants.SENDER_ID);
            String string_identifier = extras.getString(RegistrationConstants.STRING_IDENTIFIER);
            String host = extras.getString(RegistrationConstants.HOST);

            token = InstanceID.getInstance(this)
                    .getToken(sender_id, GoogleCloudMessaging.INSTANCE_ID_SCOPE, null);
            Log.d(TAG, "GCM Registration Token: " + token);

            // Register token with app server.
            boolean sent_token = sendRegistrationToServer(host, token, string_identifier);

            // You should store a boolean that indicates whether the generated token has been
            // sent to your server. If the boolean is false, send the token to your server,
            // otherwise your server should have already received the token.
            regCompleteIntent.putExtra(RegistrationConstants.SENT_TOKEN_TO_SERVER, sent_token);
        } catch (Exception e) {
            Log.e(TAG, "Failed to complete token refresh", e);
            // If an exception happens while fetching the new token or updating our registration
            // data on a third-party server, this ensures that we'll attempt the update at a later
            // time.
            regCompleteIntent.putExtra(RegistrationConstants.SENT_TOKEN_TO_SERVER, false);
        }

        Log.d(TAG, "Sending the broadcast");
        regCompleteIntent.putExtra(RegistrationConstants.EXTRA_KEY_TOKEN, token);
        LocalBroadcastManager.getInstance(this).sendBroadcast(regCompleteIntent);
    }

    /**
     * Register a GCM registration token with the app server
     * @param host App server host with port. Example "192.168.59.103:4260
     * @param token Registration token to be registered
     * @param string_identifier A human-friendly name for the client
     * @return true if request succeeds
     * @throws IOException
     */
    private boolean sendRegistrationToServer(String host, String token, String string_identifier) throws IOException {
        Uri.Builder builder = new Uri.Builder();
        builder.scheme("http")
                .authority(host)
                .appendPath("clients");

        String json = String.format(
                "{ \"registration_token\": \"%s\", \"string_identifier\": \"%s\"}",
                token, string_identifier);

        int code = GCMPlaygroundUtil.post(builder.build().toString(), json);
        return code == RegistrationConstants.VALID_POST_RESPONSE;
    }

}
