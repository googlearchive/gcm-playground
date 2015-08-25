// Copyright Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import UIKit

class ViewController: UIViewController, UITextFieldDelegate {

  let registerNewClient: String = "register_new_client"
  let unregisterClient: String = "unregister_client"

  let gcmAddress: String = "@gcm.googleapis.com"
  let topicPrefix: String = "/topics/"

  @IBOutlet weak var registrationStatus: UITextView!
  @IBOutlet weak var registrationToken: UITextView!

  @IBOutlet weak var senderIdField: UITextField!
  @IBOutlet weak var stringIdentifierField: UITextField!
  @IBOutlet weak var downstreamPayloadField: UITextView!

  @IBOutlet weak var registerButton: UIButton!
  @IBOutlet weak var unregisterButton: UIButton!

  @IBOutlet weak var topicNameField: UITextField!
  @IBOutlet weak var topicSubscribeButton: UIButton!

  @IBOutlet weak var upstreamMessageField: UITextField!
  @IBOutlet weak var upstreamMessageSendButton: UIButton!

  @IBOutlet weak var pubsubView: UIView!
  @IBOutlet weak var upstreamView: UIView!

  var apnsToken: NSData!
  var token: String = ""
  var appDelegate: AppDelegate!

  var gcmSenderID: String?
  var stringIdentifier: String?

  var registrationOptions = [String: AnyObject]()

  override func viewDidLoad() {
    super.viewDidLoad()

    appDelegate = UIApplication.sharedApplication().delegate as! AppDelegate

    // iOS registered the device and sent a token
    NSNotificationCenter.defaultCenter().addObserver(self, selector: "saveApnsToken:",
      name: appDelegate.apnsRegisteredKey, object: nil)
    // Got a new GCM reg token
    NSNotificationCenter.defaultCenter().addObserver(self, selector: "updateRegistrationStatus:",
      name: appDelegate.registrationKey, object: nil)
    // GCM Token needs to be refreshed
    NSNotificationCenter.defaultCenter().addObserver(self, selector: "onTokenRefresh:",
      name: appDelegate.tokenRefreshKey, object: nil)
    // New message received
    NSNotificationCenter.defaultCenter().addObserver(self, selector: "showReceivedMessage:",
      name: appDelegate.messageKey, object: nil)

    self.senderIdField.delegate = self
    self.stringIdentifierField.delegate = self
    self.topicNameField.delegate = self
    self.upstreamMessageField.delegate = self

    // Add borders to pubsub and upstream message views
    pubsubView.layer.borderColor = UIColor.grayColor().CGColor
    pubsubView.layer.borderWidth = 1
    pubsubView.layer.masksToBounds = true
    upstreamView.layer.borderColor = UIColor.grayColor().CGColor
    upstreamView.layer.borderWidth = 1
    upstreamView.layer.masksToBounds = false

    // TODO(karangoel): Remove this, only for development
    senderIdField.text = "1015367374593"
  }

  // Hide the keyboard when click on "Return" or "Done" or similar
  func textFieldShouldReturn(textField: UITextField) -> Bool {
    self.view.endEditing(true)
    return false
  }

  override func didReceiveMemoryWarning() {
    super.didReceiveMemoryWarning()
    // Dispose of any resources that can be recreated.
  }

  // Click handler for register button
  @IBAction func registerClient(sender: UIButton) {
    // Get the fields values
    gcmSenderID = senderIdField.text
    stringIdentifier = stringIdentifierField.text

    // Validate field values
    if (gcmSenderID == "") {
      showAlert("Invalid input", message: "Sender ID and host cannot be empty.")
      return
    }

    // Register with GCM and get token
    var instanceIDConfig = GGLInstanceIDConfig.defaultConfig()
    instanceIDConfig.delegate = appDelegate
    GGLInstanceID.sharedInstance().startWithConfig(instanceIDConfig)
    registrationOptions = [kGGLInstanceIDRegisterAPNSOption:apnsToken,
      kGGLInstanceIDAPNSServerTypeSandboxOption:true]
    GGLInstanceID.sharedInstance().tokenWithAuthorizedEntity(gcmSenderID,
      scope: kGGLInstanceIDScopeGCM, options: registrationOptions, handler: registrationHandler)
  }

  // Click handler for unregister button
  @IBAction func unregisterFromAppServer(sender: UIButton) {
    let message = ["action": unregisterClient, "token": token]
    sendMessage(message)
    token = ""
    updateUI("Unregistration COMPLETE!", registered: false)
  }


  @IBAction func topicChangeHandler(sender: UITextField) {
    var userInput = topicNameField.text
    if (userInput != "") {
      topicSubscribeButton.enabled = true
    } else {
      topicSubscribeButton.enabled = false
    }
  }

  // Got a new GCM registration token
  func updateRegistrationStatus(notification: NSNotification) {
    if let info = notification.userInfo as? Dictionary<String,String> {
      if let error = info["error"] {
        registrationError(error)
      } else if let regToken = info["registrationToken"] {
        updateUI("Registration SUCCEEDED", registered: true)
      }
    } else {
      println("Software failure.")
    }
  }

  // Show the passed error message on the UI
  func registrationError(error: String) {
    updateUI("Registration FAILED", registered: false)
    showAlert("Error registering with GCM", message: error)
  }

  // Save the iOS APNS token
  func saveApnsToken(notification: NSNotification) {
    if let info = notification.userInfo as? Dictionary<String,NSData> {
      if let deviceToken = info["deviceToken"] {
        apnsToken = deviceToken
      } else {
        println("Could not decode the NSNotification that contains APNS token.")
      }
    } else {
      println("Could not decode the NSNotification userInfo that contains APNS token.")
    }
  }

  // GCM token should be refreshed
  func onTokenRefresh() {
    // A rotation of the registration tokens is happening, so the app needs to request a new token.
    println("The GCM registration token needs to be changed.")
    GGLInstanceID.sharedInstance().tokenWithAuthorizedEntity(gcmSenderID,
      scope: kGGLInstanceIDScopeGCM, options: registrationOptions, handler: registrationHandler)
  }

  // Callback for GCM registration
  func registrationHandler(registrationToken: String!, error: NSError!) {
    if (registrationToken != nil) {
      token = registrationToken
      println("Registration Token: \(registrationToken)")
      registerWithAppServer()
    } else {
      println("Registration to GCM failed with error: \(error.localizedDescription)")
      registrationError(error.localizedDescription)
    }
  }

  func showReceivedMessage(notification: NSNotification) {
    if let info = notification.userInfo as? Dictionary<String,AnyObject> {
      downstreamPayloadField.text = info.description
    } else {
      println("Software failure. Guru meditation.")
    }
  }

  // Call the app server and register the current reg token
  func registerWithAppServer() {
    let message = ["action": registerNewClient, "token": token, "stringIdentifier": stringIdentifierField.text]
    sendMessage(message)
    self.updateUI("Registration COMPLETE!", registered: true)
    topicSubscribeButton.enabled = true
  }

  func sendMessage(message: NSDictionary) {
    // The resolution for timeIntervalSince1970 is in millisecond. So this will work
    // when you are sending no more than 1 message per millisecond.
    // To use in production, there should be a database of all used IDs to make sure
    // we don't use an already-used ID.
    let nextMessageID: String = NSDate().timeIntervalSince1970.description

    let to: String = senderIdField.text + gcmAddress
    GCMService.sharedInstance().sendMessage(message as [NSObject : AnyObject], to: to, withId: nextMessageID)
  }

  @IBAction func subscribeToTopic(sender: UIButton) {
    let topic = topicNameField.text.stringByTrimmingCharactersInSet(NSCharacterSet.whitespaceCharacterSet())
    // Topic must begin with "/topics/" and have a name after the prefix
    if (topic == "" || !topic.hasPrefix(topicPrefix) || count(topic) <= count(topicPrefix)) {
      showAlert("Invalid topic name", message: "Make sure topic is in format \"/topics/topicName\"")
      return
    }

    GCMPubSub.sharedInstance().subscribeWithToken(token, topic: topic,
      options: nil, handler: {(NSError error) -> Void in
        if (error != nil) {
          // Error 3001 is "already subscribed". Treat as success.
          if error.code == 3001 {
            println("Already subscribed to \(topic)")
            self.updateUI("Subscribed to topic \(topic)", registered: true)
          } else {
            println("Subscription failed: \(error.localizedDescription)");
            self.updateUI("Subscription failed for topic \(topic)", registered: false)
          }
        } else {
          NSLog("Subscribed to \(topic)");
          self.updateUI("Subscribed to topic \(topic)", registered: true)
        }
    })
  }

  func updateUI(status: String, registered: Bool) {
    // Set status and token text
    registrationStatus.text = status
    registrationToken.text = token

    // Button enabling
    registerButton.enabled = !registered
    unregisterButton.enabled = registered

    // Topic and upstream message field
    topicNameField.enabled = registered
    upstreamMessageField.enabled = registered
    upstreamMessageSendButton.enabled = registered
  }

  func showAlert(title:String, message:String) {
    let alert = UIAlertController(title: title,
      message: message, preferredStyle: .Alert)
    let dismissAction = UIAlertAction(title: "Dismiss", style: .Destructive, handler: nil)
    alert.addAction(dismissAction)
    self.presentViewController(alert, animated: true, completion: nil)
  }

  deinit {
    NSNotificationCenter.defaultCenter().removeObserver(self)
  }

}

