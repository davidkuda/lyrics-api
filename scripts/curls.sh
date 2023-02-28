# signup
curl \
  --header "Content-Type: application/json" \
  --request POST \
  --data '{"email":"davidkuda","password":"supersecret"}' \
  "http://localhost:8032/signup"

# signin
curl \
  --silent \ # Do not show progress bar
  -include \ # include headers in the output
  --header "Content-Type: application/json" \
  --data '{"email":"dku@dku","password":"berlin"}' \ # infer Post
  "http://localhost:8032/signin"

# parse headers
function ph {
  curl \
  --silent \
  -include \
  --header "Content-Type: application/json" \
  --data '{"email":"dku@dku","password":"berlin"}' \
  "http://localhost:8032/signin" \
  | grep "Set-Cookie:" \
  | sed "s/^Set-Cookie: //" \
  | sed "s/; /\n/g"
}
