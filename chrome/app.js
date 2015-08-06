// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the 'License');
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an 'AS IS' BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

var registrationToken = '';

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
function handleAlreadyRegistered() {
  chrome.storage.local.get('regToken', function(result) {
    registrationToken = result['regToken'];
    setStatus('Already registered. Registration token:\n' + registrationToken);
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
 * Call cb with `true` if the client is registered, false otherwise.
 */
function isRegistered(cb) {
  chrome.storage.local.get('registered', function(result) {
    cb(result['registered']);
  });
}


/**
 * Make an HTTP request.
 */
function makeRequest(method, url, body, expectedStatus, cb) {
  var xhr = new XMLHttpRequest();
  xhr.open(method, url, true);
  xhr.setRequestHeader('Content-Type', 'application/json;charset=UTF-8');
  xhr.send(body);

  xhr.onreadystatechange = function() {
    if (xhr.readyState == 4) {
      cb(xhr.status === expectedStatus, xhr.responseText);
    }
  }
}


/**
 * Register a GCM registration token with the app server.
 */
function registerWithAppServer(regToken, cb) {
  var body = JSON.stringify({registration_token: regToken });
  var url = document.getElementById('appServerHost').value + 'clients';
  makeRequest('POST', url, body, 201, function(status, err) {
    cb(status, err);
  });
}


/**
 * Unregister a registration token from the app server.
 */
function unregisterFromAppServer(cb) {
  var url = document.getElementById('appServerHost').value +
            'clients/' + registrationToken;
  makeRequest('DELETE', url, '', 204, function(status, err) {
    cb(status, err);
  });
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
  setStatus('Unregistering...');
  disableButtons();
}


/**
 * Called when GCM server responds to a registration request.
 */
function registerCallback(regToken) {
  if (chrome.runtime.lastError) {
    // Registration failed, handle the error and retry later.
    document.getElementById('register').disabled = false;
    return setStatus('FAILED: ' + chrome.runtime.lastError.message);
  }

  setStatus('Registration SUCCESSFUL. Registration ID:\n' + regToken);

  registrationToken = regToken;

  // Notify the app server about this new registration
  registerWithAppServer(registrationToken, function(succeed, err) {
    if (succeed) {
      setAppServerStatus('Registration with app server SUCCESSFUL.');
      chrome.storage.local.set({ registered: true });
      chrome.storage.local.set({ regToken: registrationToken });
      document.getElementById('register').disabled = true;
      document.getElementById('unregister').disabled = false;
    } else {
      setAppServerStatus('Registration with app server FAILED: ' + err);
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
  chrome.storage.local.remove([ 'registered', 'regToken' ]);

  // Notify the app server about this unregistration
  unregisterFromAppServer(function(succeed, err) {
    if (succeed) {
      setAppServerStatus('Unregistration with the app server SUCCESSFUL.');
      registrationToken = '';
    } else {
      setAppServerStatus('Unregistration with app server FAILED: ' + err);
    }
  });
}


window.onload = function() {
  document.getElementById('register').onclick = register;
  document.getElementById('unregister').onclick = unregister;

  chrome.gcm.onMessage.addListener(function(message) {
    console.log('chrome.gcm.onMessage', message);
  });

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
