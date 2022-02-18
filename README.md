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

# Usage

The service is meant to be used inside a Docker container. Though it should work as an executable as well, if you manage to build it.

To run it as a Docker container you should:
1. Clone the repository
2. Create ```configs``` directory with your main.yaml and filters.yaml configs.
3. run ```docker build -t janus .```
4. run ```docker run -v $(pwd)/configs:/configs/ janus /configs/main.yaml```

# Testing

There are a few integration tests which check basic functionality. You can also use them as a reference.

1. Clone the repository
2. ```cd smoke_tests```
3. ```/run_tests.sh```

# More info

- The most advanced sample config is located in 
[sample_configs/my_wirenboard_to_homeassistant.yaml](https://github.com/phoenix-mstu/janus-mqtt-proxy/blob/master/sample_configs/my_wirenboard_to_homeassistant.yaml).
Look there for more examples.

- Article in English: https://medium.com/@phoenix.mstu/modifying-mqtt-proxy-bf6d8931ef60
- Article in Russian: https://habr.com/ru/company/funcorp/blog/497234/
