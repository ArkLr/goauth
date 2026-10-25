Step 9 is mostly wiring, because you already have most of the pieces. Each session gets a random CSRFToken at login, Protected passes it to the template, and the logout form sends it as a hidden field. What's missing is the server check. At the moment Logout never reads csrf_token, so the token does nothing.

1. What you're defending against

You're logged in to GoAuth. You open evil.com in another tab, and it contains:

<form action="https://your-goauth.com/logout" method="POST"></form>
<script>document.forms[0].submit()</script>

Your browser automatically attaches your session cookie to that request, so your server thinks you clicked the button. Logout is harmless, but think of "change email", "change password" or "delete account".

There are two defences, and you'll build both:

┌──────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────────┬─────────────────────────────────────────────────────────────────────┐
│ Defence │ How it works │ What it protects │
├──────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────┤
│ Token check (Part A) │ The real form contains a secret that evil.com can't read, and the server compares it with the session's │ Forms used by logged-in users (logout and anything you add later) │
│ │ token │ │
├──────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────┤
│ Origin check (Part │ The browser says which site a request came from, and the server rejects requests from other sites │ Every POST, including login and register, where no session exists │
│ B) │ │ yet │
└──────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────┴─────────────────────────────────────────────────────────────────────┘

SameSite=Lax on your cookie already blocks most of this. These two layers are defence in ses SameSite doesn't, such as attacks from a subdomain of your own site.

2. Part A: a RequireCSRF middleware

Create it in internal/middleware/csrf.go. It runs after RequireAuth, so the session is al

The logic, in order:

1. If the method is GET, HEAD or OPTIONS, just call next. These requests must never change anything.
2. Get the session with your existing SessionFromContext. If it isn't there, that's a pro0 error.
3. Read the token the browser sent: r.PostFormValue("csrf_token"). Also fall back to the X-CSRF-Token header, which you'll want later for JavaScript fetch requests.
4. Compare it with s.CSRFToken using crypto/subtle.ConstantTimeCompare, not ==. ConstantTonvert both strings, and it returns 1 when they match.
5. If the token is empty or doesn't match, return 403 Forbidden and stop. Otherwise call next.ServeHTTP(w, r).

The function should look like this, the same pattern as RequireAuth but simpler, because it needs nothing passed in:

func RequireCSRF(next http.Handler) http.Handler

Two things to get right:

- Check that the token is not empty. Comparing an empty token with an empty token would p
- Keep the request body size limit. PostFormValue reads the body for you. Your handlers set MaxBytesReader themselves, so set it in this middleware too, before reading the form.

Wire it up in main.go:

mux.Handle("POST /logout", requireAuth(RequireCSRF(http.HandlerFunc(auth.Logout))))

The order matters: the outermost wrapper runs first, so requireAuth puts the session in the context before RequireCSRF needs it.

Simplify Logout. It's now behind requireAuth, so it can take the session from SessionFromContext instead of reading the cookie again. You'll need to delete it by its token, and remember the database stores
s.Token already hashed.

3. Part B: an origin check on every request

Go has this built in: http.CrossOriginProtection (added in Go 1.25; you're on 1.27). It cer, which every modern browser sends, and falls back to comparing Origin with Host. Itrejects POST requests from other sites.

Run go doc net/http.CrossOriginProtection and read the docs, then:

1. Create one: cop := http.NewCrossOriginProtection().
2. Look at its Handler method and wrap your whole mux with it, then pass the result to ListenAndServe.

That protects login and register too. This matters because of login CSRF: an attacker can secretly log you into their account. Anything you then save goes into their account.

If you put nginx or Caddy in front later, make sure the proxy passes the original Host header (nginx: proxy_set_header Host $host;). Otherwise the Origin/Host fallback check fails.

4. Testing it

Test each part on its own: first add only RequireCSRF and test it, then add the origin check and test that. If you add both at once, a 403 doesn't tell you which one rejected the request.

# 1. Log in and save the cookie

curl -c jar.txt -d "email=you@x.com&password=yourpass" localhost:8080/login

# 2. Logout without a token → expect 403

curl -b jar.txt -i -X POST localhost:8080/logout

# 3. Logout with a wrong token → expect 403

curl -b jar.txt -i -d "csrf_token=nope" localhost:8080/logout

# 4. Real browser: click "Sign out" → you should end up on /login

Test the origin check with a request that claims to come from another site: curl -b jar.tss-site" -d "csrf_token=..." localhost:8080/logout. You should get 403 even with thecorrect token. Plain curl requests without that header are allowed through, by design, because they aren't coming from a browser.

Simulate a real attack: save the evil.com form from section 1 as evil.html, serve it on a different port (python3 -m http.server 9000), and open it while you're logged in. Your session should survive.

Write it as a Go test: use httptest.NewRecorder() and httptest.NewRequest, and wrap a dummy handler with RequireCSRF. Put a fake session in the context and check you get 403 without a token and 200 with the
right one. This is a good first test to write, because it doesn't need a database.

5. Two other things

- Your server ignores the PORT setting. You build addr from it, but then call ListenAndSeinstead. Note that addr uses 127.0.0.1, so the server will only accept connections fromthe same machine. That's right if a proxy sits in front of it, but if you open the port directly from another machine, the connection will fail.
- Rule for the future: never change anything in a GET handler. Both defences above only per unsafe methods).

When you've written RequireCSRF, paste it here and I'll review it.
