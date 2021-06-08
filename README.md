# OpenSlides-ICC

The Autoupdate Service is part of the OpenSlides environment. Clients can
connect to it and communicate with eachother.

IMPORTANT: The data are sent via an open http-connection. All browsers limit the
amount of open http1.1 connections to a domain. For this service to work, the
browser has to connect to the service with http2 and therefore needs https.
