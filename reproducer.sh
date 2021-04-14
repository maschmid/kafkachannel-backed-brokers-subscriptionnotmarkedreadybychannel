#!/usr/bin/env bash

make apply

# Wait for Readiness
kubectl wait --for=condition=Ready broker broker --timeout=60s
kubectl wait --for=condition=Ready ksvc counter --timeout=60s
kubectl wait --for=condition=Ready ksvc sender --timeout=60s
kubectl wait --for=condition=Ready trigger counter --timeout=60s

senderUrl=$(kubectl get ksvc sender -o template='{{.status.url}}')

echo "Sleeping 10s for ingress"
sleep 10

# Invoke the sender and immediately delete the Broker.
# The `sender` sends event to `counter` which repeatedly replies with events
# so it is very likely some events are still flowing while the Broker is being deleted.
# (which is what seem to trigger the issue)
curl ${senderUrl} && kubectl delete broker broker

# Also delete the rest of it
kubectl delete -f config/

# Now sleep for a minute, to demonstrate that new Subscriptions won't become Ready 
# even after a while after the broken deployment is deleted
echo "Sleeping for a minute"
sleep 60

# Do another make apply, we create the same resources, just with a different name 
# (re-creating the same resources would "fix" the problem)
make apply-second-part

# Watch the Subscription, see if it becomes Ready
watch kubectl get subscriptions
