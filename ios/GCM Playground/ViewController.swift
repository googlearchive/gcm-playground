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
  let statusRegistered: String = "registered"
  let statusUnegistered: String = "unregistered"

  let keyAction: String = "action"
  let keyStatus: String = "status"
  let keyRegistrationToken: String = "registration_token"
  let keyStringIdentifier: String = "stringIdentifier"

  let gcmAddress: String = "@gcm.googleapis.com"
  let topicPrefix: String = "/topics/"

  @IBOutlet weak var registrationStatus: UITextView!
  @IBOutlet weak var registrationToken: UITextView!

  @IBOutlet weak var senderIdField: UITextField!
  @IBOutlet weak var stringIdentifierField: UITextField!
  @IBOutlet weak var downstreamPayloadField: UITextView!

  @IBOutlet weak var registerButton: UIButton!
  @IBOutlet weak var unregisterButton: UIButton!
  @IBOutlet weak var progressIndicator: UIActivityIndicatorView!

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
    NSNotificationCenter.defaultCenter().addObserver(self, selector: "handleReceivedMessage:",
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

    senderIdField.text = "<your_sender_ID>"
    self.stringIdentifierField.text = "<a_name_to_recognize_the_device>"
  }

  // Hide the keyboard when click on "Return" or "Done" or similar
  func textFieldShouldReturn(textField: UITextField) -> Bool {
    self.view.endEditing(true)
    return false
  }

  override func didReceiveMemoryWarning() {
    super.didReceiveMemoryWarning()
  }

  ////////////////////////////////////////
  // Actions (Listeners)
  ////////////////////////////////////////

  // Register button click handler.
  @IBAction func registerClient(sender: UIButton) {
    // Get the fields values
    gcmSenderID = senderIdField.text
    stringIdentifier = stringIdentifierField.text

    // Validate field values
    if (gcmSenderID == "") {
      showAlert("Invalid input", message: "Sender ID and host cannot be empty.")
      return
    }

    progressIndicator.startAnimating()

    // Register with GCM and get token
    var instanceIDConfig = GGLInstanceIDConfig.defaultConfig()
    instanceIDConfig.delegate = appDelegate
    GGLInstanceID.sharedInstance().startWithConfig(instanceIDConfig)
    registrationOptions = [kGGLInstanceIDRegisterAPNSOption:apnsToken,
      kGGLInstanceIDAPNSServerTypeSandboxOption:true]
    GGLInstanceID.sharedInstance().tokenWithAuthorizedEntity(gcmSenderID,
      scope: kGGLInstanceIDScopeGCM, options: registrationOptions, handler: registrationHandler)
  }

  // Unregister button click handler.
  @IBAction func unregisterFromAppServer(sender: UIButton) {
    let message = [keyAction: unregisterClient, keyRegistrationToken: token]
    progressIndicator.startAnimating()
    sendMessage(message)
  }

  // Topic field editing handler.
  @IBAction func topicChangeHandler(sender: UITextField) {
    var userInput = topicNameField.text
    if (userInput != "") {
      topicSubscribeButton.enabled = true
    } else {
      topicSubscribeButton.enabled = false
    }
  }

  // Send upstream message button handler.
  @IBAction func sendUpstreamMessage(sender: UIButton) {
    let text = upstreamMessageField.text
    if (text == "") {
      showAlert("Can't send message", message: "Please enter a message to send")
      return
    }

    let message = ["message": text]
    sendMessage(message)
    showAlert("Message sent successfully", message: "")
  }

  // Subscribe to topic button handler.
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

  ////////////////////////////////////////
  // Utility functions
  ////////////////////////////////////////

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

  // Handles a new downstream message.
  func handleReceivedMessage(notification: NSNotification) {
    progressIndicator.stopAnimating()
    if let info = notification.userInfo as? Dictionary<String,AnyObject> {
      if let action = info[keyAction] as? String {
        if let status = info[keyStatus] as? String {
          // We have an action and status
          if (action == registerNewClient && status == statusRegistered) {
            self.updateUI("Registration COMPLETE!", registered: true)
            topicSubscribeButton.enabled = true
            return
          } else if (action == unregisterClient && status == statusUnegistered) {
            token = ""
            updateUI("Unregistration COMPLETE!", registered: false)
            return
          }
        }
      }
      downstreamPayloadField.text = info.description
    } else {
      println("Software failure. Guru meditation.")
    }
  }

  // Calls the app server to register the current reg token.
  func registerWithAppServer() {
    let message = [keyAction: registerNewClient, keyRegistrationToken: token, keyStringIdentifier: stringIdentifierField.text]
    sendMessage(message)
  }

  // Sends an upstream message.
  func sendMessage(message: NSDictionary) {
    // The resolution for timeIntervalSince1970 is in millisecond. So this will work
    // when you are sending no more than 1 message per millisecond.
    // To use in production, there should be a database of all used IDs to make sure
    // we don't use an already-used ID.
    let nextMessageID: String = NSDate().timeIntervalSince1970.description

    let to: String = senderIdField.text + gcmAddress
    GCMService.sharedInstance().sendMessage(message as [NSObject : AnyObject], to: to, withId: nextMessageID)
  }

  ////////////////////////////////////////
  // UI functions
  ////////////////////////////////////////

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

  // Shows a toast with passed text.
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

