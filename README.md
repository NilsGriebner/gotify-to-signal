# gotify-to-signal 

This Gotify plugin forwards received messages to the Signal messenger.

## Prerequisites 

>[!IMPORTANT]
To proceed with the installation, you need the [Signal Messenger REST API](https://github.com/bbernhard/signal-cli-rest-api)
container and a spare phone number. If you have that in place, follow the steps below.

To make the plugin work, you need a running
[Signal Messenger REST Api](https://github.com/bbernhard/signal-cli-rest-api) instance. Refer to the 
[getting started section](https://github.com/bbernhard/signal-cli-rest-api?tab=readme-ov-file#getting-started) on how to
run the container.

The Signal Messenger REST Api needs to be reachable from the Gotify server.

If you have a working Signal Messenger REST Api, you can register a spare phone number which will be used to send the 
messages received by Gotify. (It is also possible to link your primary phone number to the Signal Messenger REST Api, but
this setup is currently untested.)

To register your phone number, you need to generate a Signal captcha at their
[website](https://signalcaptchas.org/registration/generate). Confirm you're a humand and then grab the captcha string
from the `Open Signal` link. The string looks like `signal-hcaptcha.<random stuff>.registration.<random stuff>`.

Now you can call the REST Api with the following command:
```bash
curl -X POST -H "Content-Type: application/json" 'https://<your host>/v1/register/<your phone number in international number format>' \                                                                                  (production/ngrstack)
     -d '{"captcha": "<your generated captcha>", "use_voice": false}'
````

Afterward you receive a text response with a confirmation code. Grab the code and call the REST Api again with the 
following command:

```bash
curl -X POST -H "Content-Type: application/json" 'https://<your host>/v1/register/<your phone number>/verify/<your confirmation code>' 
````

Now you're ready to use the registered phone number with the Signal Messenger REST Api. Send a test message to verify 
that everything works as expected.

```bash
curl -X POST -H "Content-Type: application/json" 'https://<your host>/v2/send' \                                                                                                    (production/ngrstack)
     -d '{"message": "Test via Signal API!", "number": "<your registered number>", "recipients": [ "<number to send to>" ]}'
````

You can now proceed with the installation of the plugin.

## Installation :rocket:

If you already have a running Gotify server, generate a new client token for the plugin. After that make sure
all necessary environment variables are set in your deployment (you can also skip this step and configure the plugin 
later in the Gotify web interface, but it's recommended to set them beforehand).

```bash
SIGNAL_FROM_NUMBER=<your registered number in international number format>
SIGNAL_TO_NUMBER=<the number to forward messages to (in international format)>
SIGNAL_API_HOST=<your signal api host like https://signal-api.mydomain.tld>
GOTIFY_HOST=<Gotify web api host like ws://localhost:8080>
SIGNAL_CLIENT_TOKEN=<Gotify client token generated above>
```

Now you can deploy the plugin by [building](#building-hammer_and_wrench) it yourself or grab a binary release from the release
section of this repository.

## Building :hammer_and_wrench:

For building the plugin, gotify/build docker images are used to ensure compatibility with 
[gotify/server](https://github.com/gotify/server).

Export `GOTIFY_VERSION` and set it to a tag, commit or branch from the gotify/server repository.

```bash
export GOTIFY_VERSION=v2.8.0
```

This command builds the plugin for amd64, arm-7 and arm64. 
The resulting shared object will be compatible with gotify/server version 2.8.0.

```bash
make build
```
