// Copyright 2015 Google Inc. All Rights Reserved.
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

/**
 * Sets the status of the app to the passed text.
 */
function setStatus(text) {
  document.getElementById('status').value = text;
}


/**
 * Sets the status of the app server.
 */
function setAppServerStatus(text) {
  document.getElementById('appServerStatus').value = text;
}


/**
 * Handles the case when the client is already registered.
 */
function handleAlreadyRegistered(regId) {
  chrome.storage.local.get('regId', function(result) {
    setStatus('Already registered. Registration Id:\n' + result['regId']);
    document.getElementById('register').disabled = true;
    document.getElementById('unregister').disabled = false;
  });
}


/**
 * Disable the buttons to prevent multiple clicks.
 */
function disableButtons() {
  document.getElementById('register').disabled = true;
  document.getElementById('unregister').disabled = true;
}


/**
 * Call the passed cb with `true` if the client is registered, false otherwise.
 */
function isRegistered(cb) {
  chrome.storage.local.get('registered', function(result) {
    cb(result['registered']);
  });
}


/**
 * Register the passed registrationId with our app server.
 */
function sendRegistrationId(registrationId, cb) {
  // TODO(karangoel): actually implement this
  console.log('Pretending to make a HTTP request.');
  var appServerHost = document.getElementById('appServerHost').value;
  console.log('Pretending like it succeeded');
  cb(true);
}


/**
 * Calls the GCM API to register this client if not already registered.
 */
function register() {
  isRegistered(function(registered) {
    // If already registered, bail out.
    if (registered) return handleAlreadyRegistered();

    var senderId = document.getElementById('senderId').value;
    if (!senderId) {
      return setStatus('Please provide a valid sender ID.');
    }
    chrome.gcm.register([senderId], registerCallback);
    setStatus('Registering...');
    disableButtons();
  });
}


/**
 * Calls the GCM API to unregister this client.
 */
function unregister() {
  chrome.gcm.unregister(unregisterCallback);
  // TODO(karangoel): also unregister from the server
  setStatus('Unregistering...');
  disableButtons();
}


/**
 * Called when GCM server responds to a registration request.
 */
function registerCallback(registrationId) {
  if (chrome.runtime.lastError) {
    // Registration failed, handle the error and retry later.
    document.getElementById('register').disabled = false;
    return setStatus('FAILED: ' + chrome.runtime.lastError.message);
  }

  setStatus('Registration SUCCESSFUL. Registration ID:\n' + registrationId);

  // Notify the app server about this new registration
  sendRegistrationId(registrationId, function(succeed) {
    if (succeed) {
      setAppServerStatus('Registration with app server COMPLETED.');
      chrome.storage.local.set({ registered: true });
      chrome.storage.local.set({ regId: registrationId });
      document.getElementById('register').disabled = true;
      document.getElementById('unregister').disabled = false;
    } else {
      // TODO(karangoel): Show why it failed (passed in the callback)
      setAppServerStatus('Registration with app server FAILED.');
      document.getElementById('register').disabled = false;
    }
  });
}


/**
 * Called when GCM server responds to an unregistration request.
 */
function unregisterCallback() {
  document.getElementById('register').disabled = false;
  document.getElementById('unregister').disabled = true;

  if (chrome.runtime.lastError) {
    return setStatus('FAILED: ' + chrome.runtime.lastError.message);
  }

  setStatus('Unregistration SUCCESSFUL');
  chrome.storage.local.remove([ 'registered', 'regId' ]);
}


window.onload = function() {
  document.getElementById('register').onclick = register;
  document.getElementById('unregister').onclick = unregister;

  isRegistered(function(registered) {
    if (registered) {
      handleAlreadyRegistered();
    } else {
      setStatus('Put your sender ID above and click Register.');
      document.getElementById('register').disabled = false;
      document.getElementById('unregister').disabled = true;
    }
  });

};
