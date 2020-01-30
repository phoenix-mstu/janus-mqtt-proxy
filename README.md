# Modifying MQTT proxy

With this service you can transform MQTT topics and their payload 
by specifying transformation rules.

For example you have topics `/some/topic1` and `/some/another/topic2` on MQTT broker, 
and you have a client which needs them to be `/root/topic1` and `/root/topic2`. 

All you need is to create a yaml config with this rules:

```yaml

broker_to_client:

  - topic: ^/some/topic1$
    template: '/root/topic1'
    val_map: {0: OFF, 1: ON}        # you can specify map to transform values
  
  - topic: ^/some/another/([^/]*)$  # you can use regex
    template: '/root/{{.f1}}'       # and place an extracted part into template


client_to_broker:

  - topic: ^/root/topic1$
    template: '/some/topic1'
    val_map: {OFF: 1, ON: 1}        # reverse value transformation

  # if you comment it, this topic will be readonly for client
  # - topic: ^/root/topic2$
  #   template: '/some/another/topic2'
```

And another yaml file with basic info:

```yaml
broker_host: tcp://broker.host:1883
broker_login: login
broker_password: pass

clients:

  - login: some_login
    password: some_pass
    filters_config: your_filters.yaml
  
  # another client with no password
  - filters_config: another_filters.yaml
```

Then you can run proxy: `./proxy basic.yaml`

The most advanced sample config is located in 
[sample_configs/my_wirenboard_to_homeassistant.yaml](https://github.com/phoenix-mstu/go-modifying-mqtt-proxy/blob/master/sample_configs/my_wirenboard_to_homeassistant.yaml).
Look there for more examples.