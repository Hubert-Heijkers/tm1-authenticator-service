# Authenticator Service

The "Authenticator" service is a very simple, sample HTTP service that implements HTTP authentication as defined in [RFC7235](https://datatracker.ietf.org/doc/html/rfc7235). This sample is primarily intended to provide the plumbing required to offer a service that can be used in combination with TM1 v12 set up for HTTP pass-through authentication. The actual authentication logic will need to be implemented; the sample merely checks if the username provided is an e-mail address in the `example.com` domain and if the password is equal to the well-known magic "apple" password.

## The Service

This service supports an "ActiveUser" resource which can be invoked using a GET request. The request MUST contain the `Authorization` header specifying the credentials required to authenticate with the service. If authentication fails, it will return a `401 Unauthorized` and a `WWW-Authenticate` header identifying the form of authentication required. If authentication is successful, it will return a `200 OK`, and the body will contain a JSON object with a `Name` property and the value containing the subject identifier, in this case, the e-mail address from the "example.com" domain used to authenticate with the service.

Note that the returned contents follow the structure that a TM1 v12 service requires when using this service for HTTP pass-through authentication, which in turn is compatible with the `User` type of TM1's OData compliant REST API. However, it doesn't need to do so; any structure that suits your implementation would be fine as long as it contains the name, a.k.a. Subject ID or Subject Name, of the authenticated user, and the property holding that name can be identified using a JSON pointer (see TM1 v12's authentication documentation for more information).

### Example Requests

#### Requesting the details about the ActiveUser

When requesting the ActiveUser from a browser as in:

```url
http://localhost:8080/ActiveUser
```

You'll see that the browser will, at least the first time (clear cache after each use), ask for your username and password as it will have gotten a `401 Unauthorized` response from our authenticator service. After you have entered the appropriate username, an e-mail address from the "example.com" domain, and a password, you know which one, it'll retry but this time injecting the `Authorization` header containing the credentials, after which the service will return the information, the name, about the authenticated user.

When called from the command line using curl:

```bash
curl "http://localhost:8080/ActiveUser" \
    -H "Authorization: Basic dXNlckBleGFtcGxlLmNvbTphcHBsZQ=="
```

You'd pass in the `Authorization` header yourself, identifying the authentication method, Basic in this example, and the base-64 encoded `username:password`, again resulting in the service returning a JSON object containing the name of the authenticated user.

### How to use

Run the Go application directly from source, optionally specifying a port using the --port flag, as in:

```bash
go run main.go --port=9090
```

This will start the Authenticator service listening to port `9090`. If no port is specified, it will default to port `8080`.

Once you are satisfied with the service you would first build an executable using:

```bash
go build
./tm1-authenticator-service.exe --port=9090
```

This places the executable in the root of the source directory, or:

```bash
go install
tm1-authenticator-service.exe --port=9090
```

This also builds the executable but instead of cluttering your source directory, it places it in the `bin` folder of your workspace defined by the `GOPATH` environment variable. This snippet assumes that the `bin` folder is in your path, thus the OS will know where to find your executable for the Authenticator service.

### Installing the Authenticator service

The code checks if it is running as a Windows service and will act accordingly. To set up the Authenticator service as a Windows service, create/register the service and start it with sc.exe:

```bash
sc.exe create TM1-Authenticator-Service binPath= "C:\path\to\your\tm1-authenticator-service.exe"
sc.exe start TM1-Authenticator-Service
```

## Using the Authenticator service with TM1 v12

While this service could potentially be generally useful, provided you implemented a solid mechanism to perform the actual authentication of the user and/or use it as a wrapper for an already existing authentication method you want to make available through HTTP authentication, it was written for demo purposes wherein a simple HTTP authentication provider was required to provide a set of authenticated users (in this example, anybody in the "example.com" domain).

At the time of this writing, HTTP pass-through authentication was only available in the (beta) version of the standalone/local TM1 v12 service. See the TM1 v12 documentation for more information about using HTTP pass-through authentication and how to configure it.

> Did you know that the idea for HTTP pass-through authentication came from one of the largest TM1 customers? It allowed them to point a v12 TM1 service at a v11 server, in their case configured for LDAP, allowing them to continue using their existing authentication infrastructure. And whilst CAM doesn't strictly follow the HTTP authentication specification, not even close, pointing a v12 TM1 service to a v11 server will even allow you to continue using CAM for the time being as well!
