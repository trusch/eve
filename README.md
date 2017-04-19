Eve - Eve is a Virtual Entrypoint
=================================

Eve is a virtual entrypoint for your micro service deployment. It is a thin layer on top of etcd which handles HTTP and HTTPS requests and routes them to your backend containers.

## Basics
Eve is very simple: There are only three concepts you need to understand to use eve effectively:
### Loadbalancer Rules
Eve maps requests to loadbalancers. To do so you can specify loadbalancer rules.
A rule consists of the following parts:
* ID: a unique rule name
* Route: a routing-expression which decides whether this rule applies to a requests
  * The routing language is taken from [vulcand-route](https://github.com/vulcand/route)
  * Example route: `Host("echo.mydomain.tld") && Path("/v1")`
* Target: a loadbalancer ID to map this request

### Loadbalancer Hosts
If a requests maps to a specific loadbalancer, eve must know about backendservices serving the request.
Therefore a loadbalancer has hosts associated with it.

### Middleware Rules
Eve can additionally apply middleware to the request handler chain. A middleware rule is structured like a loadbalancer rule, but doesn't specify which loadbalancer should be targeted, but rather which middlewares should be applied to the request.
So a rule consists of the following parts:
* ID: a unique rule name
* Route: a routing-expression which decides whether this rule applies to a requests
* Middlewares: an array of middleware objects.
  * a middleware object contains an ID specifying the middleware type and Opts field which is passed to the middleware constructor.


## Usage
### Start and configure eve
#### With rkt
Since eve's main config service is etcd, you need to run etcd first
```bash
sudo rkt run --net=host trusch.io/etcd
```
Then you can start eve:
```bash
sudo rkt run --net=host trusch.io/eve
```
Now eve is up and running and you can start services, for example:
```bash
sudo rkt run --net=default:IP=172.16.28.2 trusch.io/http-echo
```
Lets configure eve!
```bash
sudo rkt run --net=host trusch.io/eve-ctl -- \
  loadbalancer rule add \
    --id echo-lb-rule \
    --route 'Host("echo.mydomain.tld")' \
    --target echo-lb

sudo rkt run --net=host trusch.io/eve-ctl -- \
  loadbalancer host add \
    --id echo-worker-1 \
    --loadbalancer echo-lb \
    --url http://172.16.28.2
```

If everything went well, we can now open our browser and open `http://echo.mydomain.tld`. If the DNS is configured properly
so that the URL points to our deployment, we should see the response of the http-echo service.

#### With docker
Eve can also be used with docker. Besides using the approach from above (etcd + eve + http-echo + manual configure) eve can be configured to listen for docker events.
```bash
docker run -v /var/run/docker.sock:/var/run/docker.sock --net host -it trusch/eve eve --docker
```
Now start a service with a specific label:
```bash
docker run --label "eve.host=echo.mydomain.tld" trusch/http-echo
```
Eve recognizes the label and creates a loadbalancer rule with route Host("echo.mydomain.tld") and adds the container as host to the loadbalancer this rule points to. Like above if everything went well, we can now open our browser and open `http://echo.mydomain.tld`. If the DNS is configured properly so that the URL points to our deployment, we should see the response of the http-echo service.

### Use Middleware
To apply some middleware, just configure a middleware rule:
```bash
sudo rkt run --net=host trusch.io/eve-ctl -- \
  middleware rule add \
    --id echo-mw-rule \
    --route 'Host("echo.mydomain.tld")' \
    --middleware '[{"id": "trace", "opts":{"output":"/dev/stderr"}}]'
```

### Use HTTPS
Eve can terminate HTTPS requests for you. Just supply certificates and eve will choose the correct one automatically with SNI. Certificates are stored AES encrypted in etcd, so you must supply the same `--password` here and in your eve-start-command.
```bash
sudo rkt run --net=host trusch.io/eve -- --password super-secure-password
sudo rkt run \
  --volume certs,kind=host,source=/etc/certs --net=host \
  trusch.io/eve-ctl --mount volume=certs,target=/etc/certs -- \
    cert add \
      --cert /etc/certs/echo.crt \
      --key /etc/certs/echo.key \
      --password super-secure-password
```
