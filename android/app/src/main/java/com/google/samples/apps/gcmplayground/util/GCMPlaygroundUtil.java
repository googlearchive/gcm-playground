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

package com.google.samples.apps.gcmplayground.util;

import android.util.Log;

import com.squareup.okhttp.MediaType;
import com.squareup.okhttp.OkHttpClient;
import com.squareup.okhttp.Request;
import com.squareup.okhttp.RequestBody;
import com.squareup.okhttp.Response;

import java.io.IOException;

public class GCMPlaygroundUtil {

    private static final MediaType JSON = MediaType.parse("application/json;charset=utf-8");
    private static final String TAG = "GCMPlaygroundUtil";
    private static final int ERROR_CODE_INT = -1;

    /**
     * Make a post request.
     * @param url URL to POST to
     * @param json JSON data to send
     * @return status code returned for the request
     */
    public static int post(String url, String json) {
        Log.d(TAG, "Making POST request: " + url);
        Log.d(TAG, json);

        OkHttpClient client = new OkHttpClient();
        RequestBody body = RequestBody.create(JSON, json);
        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();
        Response response = null;
        try {
            response = client.newCall(request).execute();
            return response.code();
        } catch (IOException e) {
            Log.e(TAG, "Error making request to " + url, e);
            return ERROR_CODE_INT;
        }
    }

    /**
     * Make a delete request.
     * @param url URL to DELTE to
     * @return status code returned for the request
     */
    public static int delete(String url) {
        Log.d(TAG, "Making DELETE request: " + url);

        OkHttpClient client = new OkHttpClient();
        RequestBody body = RequestBody.create(JSON, "");
        Request request = new Request.Builder()
                .url(url)
                .delete(body)
                .build();
        Response response = null;
        try {
            response = client.newCall(request).execute();
            return response.code();
        } catch (IOException e) {
            Log.e(TAG, "Error making request to " + url, e);
            return ERROR_CODE_INT;
        }
    }

}

