# spotify-slack
## Running this
1. Get yourself a spotify client ID and token. Go to https://developer.spotify.com, login with your spotify account and then head to dashboard, and click "create client ID"
1. Set up a slackbot in [your apps](https://api.slack.com/apps), head to "oauth & permissions" and add the chat:write scope and the commands scope
1. Click "install to workspace", head to "Basic information", scroll down and copy the verification token
1. Run `SPOTIFY_ID=<ID> SPOTIFY_SECRET=<SECRET> SLACK_VERIFICATION_TOKEN=<TOKEN> go run main.go`, then go to the URL shown in the terminal window to handle oauth with spotify.
1. Install [ngrok](https://ngrok.com), add it to your path, then in a separate terminal window, run `ngrok http 8080` and copy the url in "forwarding". 
1. Head to "slash commands" back in the slack setup, add a new one for "/spotify" or similar and set the request url to the URL you just copied
1. Run a slash command (I do this in messages to myself) `/spotify add <query>` and it should give you back a list of songs!