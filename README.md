# kafkachannel-backed-brokers-subscriptionnotmarkedreadybychannel
Reproducer for Knative Eventing Kafka issue, Events flowing after Broker delete causes new Subscriptions to fail

Prerequisites:
* Make sure `$GOPATH/bin/ko` exists

Check ./reproducer.sh and Makefile if they make sense for your env.

Run

```
./reproducer.sh
```

If the problem occurs, you'd see something like

```
NAME                                                              AGE   READY     REASON
broker-part-two-counter-part-tw13bd404dc66a9961a160cbd8ccba49bd   10m   Unknown   SubscriptionNotMarkedReadyByChannel
broker-part-two-counter-part-tw97e43bd85735d9a54550abf11faac924   10m   Unknown   SubscriptionNotMarkedReadyByChannel
```

With the new subscriptions hang in SubscriptionNotMarkedReadyByChannel for tens of minutes.
