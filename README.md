# Release-finder Project


### App To Get Latest Versions Of Github Apps And Expose As Gauge Prometheus Metrcs 


## How To Use

add new services with its address like below into configmap/configs.yml

```

if it has releases:
https://api.github.com/{ repository }/releases/latest

eg:
  - service: elasticsearch
    url: https://api.github.com/repos/elastic/elasticsearch/releases/latest

if it has tags:
https://api.github.com/{ repository }/tags

eg:
  - service: keepalived
    url: https://api.github.com/repos/acassen/keepalived/tags


change configmap with new services.

```