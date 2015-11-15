# Metrik

[![GoDoc](https://godoc.org/github.com/michaelbironneau/metrik?status.svg)](https://godoc.org/github.com/michaelbironneau/metrik)

**Warning: This is experimental, use at your own risk!**

Metrik is a small, embeddable library with *no dependencies* that makes it easy to expose one or more real-time metrics through a JSON API. Metric values are points that can be tagged with key-value pairs, such as "region: europe". Metrik indexes these points in memory and allows your users to run aggregated and group-by queries against them. Metrik handles all the routing and comes with pluggable hooks for middleware such as authentication.

At Open Energi, we use Metrik to power our real-time [public API](https://jsapi.apiary.io/previews/oerealtimeapi/reference) and our internal real-time API.

**Types of queries that a Metrik-powered API could answer**

* Total aggregate (eg. `SELECT AVG(cpu) WHERE rack_id = 132`)
* Group by aggregate (eg. `SELECT COUNT(server) WHERE tenant = 'tech-startup123' GROUP BY datacenter`)

Here are some queries that we are using Metrik to answer:

* Total power consumption of assets we control, grouped by UK region, for a certain asset type
* Total energy flexibility of assets we control

**What you provide**

* A list of metrics (a bit of metadata to identify each metric and an updater that runs in a goroutine)
* A list of tag keys (so that users can see available tag keys). Tag values can be dynamic. 

**What Metrik does with that**

* It starts your updater and captures any updates you send it via a channel
* It indexes this data in an in-memory inverted index structure
* It serves an HTTP API for consumers to slice and dice this data (group-by aggregation)

**Optional extras**

* Custom aggregators
* Custom logging
* Custom authentication providers
* Pluggable interface to transform/enrich the response

## Metrik's HTTP/JSON API

This is what your users will see. For an example of a real API that is powered by Metrik, check out the Apiary docs for our [public real-time API](https://jsapi.apiary.io/previews/oerealtimeapi/reference).

There are four routes:

* `/metrics`: List of metrics and their metadata
* `/tags`: List of tag groups and their metadata (eg. `{"name": "region", "description": "UK region (NUTS 1)"})`)
* `/:aggregate/:metric[?tag_1=val_1[&tag_2=val_2[&...tag_n=val_n]]]`: Total aggregate with optional filtering. The equivalent SQL would be `SELECT :aggregate(:metric) WHERE tag_1 = val_1 AND tag_2 = val_2 AND ... tag_n = val_n`. For example `sum/memory/?app=blog`.
* `/:aggregate/:metric/by/:tag[?tag_1=val_1[&tag_2=val_2[&...tag_n=val_n]]]`: group by aggregate with optional filtering. The equivalent SQL would be `SELECT :aggregate(:metric) WHERE tag_1 = val_1 AND tag_2 = val_2 AND ... tag_n = val_n GROUP BY :tag`. For example `count/server/by/tenant`.

There are three built-in aggregates: `count`, `sum`, and `average`. It is easy to add your own by implementing the Aggregator interface.

Here is an example query and response pair:

```
GET /average/cpu/by/rack

{
    "metrics": [
        {
            "name": "cpu",
            "groups": [
                {
                    "key": "2",
                    "value": 0.6933896206777048
                },
                {
                    "key": "0",
                    "value": 1.3587985800806675
                },
                {
                    "key": "1",
                    "value": 1.1377826548716587
                }
            ]
        }
    ]
}
```

## Tutorial

Coming soon. For now check out the code sample in  the `example` folder.