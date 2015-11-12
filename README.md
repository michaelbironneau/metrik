# Metrik

**Warning: This is very experimental and not very well tested. Use at your own peril**

Metrik is a small library to help you expose one or more real-time metrics through a JSON API. Metric values are points that can be tagged with key-value pairs, such as "region: europe". 

**What you provide**

* A list of metrics (a bit of metadata to identify each metric and an updater that runs in a goroutine)
* A list of tag keys (so that users can see available tag keys). Tag values can be dynamic. 

**What Metrik does with that**

* It starts your updater and captures any updates you send it via a channel
* It indexes this data in an in-memory inverted index structure
* It serves an HTTP API for consumers to slice and dice this data (group-by aggregation)

## The slice-and-dice HTTP API

More documentation on this coming soon. Here's an example of a request-response pair.


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

