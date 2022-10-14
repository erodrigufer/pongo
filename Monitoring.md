# Notes on Monitoring

## Table of contents

<!-- vim-markdown-toc GitLab -->

* [References](#references)
* [Motivation for monitoring](#motivation-for-monitoring)
* [Black-Box Versus White-Box](#black-box-versus-white-box)
* [The Four Golden Signals](#the-four-golden-signals)
	- [Application-specific signalling](#application-specific-signalling)
	- [Mean latency](#mean-latency)
	- [Status of served traffic (HTTP)](#status-of-served-traffic-http)
* [Granularity of measurements](#granularity-of-measurements)
* [Versioning, rollouts and monitoring](#versioning-rollouts-and-monitoring)
	- [Monitoring system configuration](#monitoring-system-configuration)
* [Domain-specific monitoring (Go and Docker)](#domain-specific-monitoring-go-and-docker)
* [Prometheus](#prometheus)
	- [Metrics](#metrics)
	- [Further information and guides](#further-information-and-guides)

<!-- vim-markdown-toc -->

## References
1. [Site Reliability Engineering: How Google runs production systems](https://sre.google/sre-book/monitoring-distributed-systems/): Chapter 6 Monitoring distributed systems.
2. [The Site Reliability Workbook: Practical ways to implement SRE](https://sre.google/workbook/monitoring/): Chapter 4 Monitoring.
3. [How Prometheus monitoring works](https://www.youtube.com/watch?v=h4Sl21AKiDg): An introductory video explaining the architecture and functionality of Prometheus.
4. [Available . . . or not? That is the question—CRE life lessons](https://cloud.google.com/blog/products/gcp/available-or-not-that-is-the-question-cre-life-lessons): Google Cloud blog post about how to practically define availability with black-box testing and how to choose availability targets with a business/product-driven approach.
5. [Understanding metric types (Prometheus)](https://prometheus.io/docs/tutorials/understanding_metric_types/): The official Prometheus documentation explains the four different metric types.
6. [Collecting Prometheus metrics in Golang](https://gabrieltanner.org/blog/collecting-prometheus-metrics-in-golang/): A very detailed guide about collecting metrics with Golang using Prometheus and finally displaying them in a dashboard with Grafana.

## Motivation for monitoring
1. Alert on critical system conditions.
2. Help to investigate and hopefully fix service availability issues.
3. Display information about the system visually (dashboard).
4. Gain statistical insights into service health and resource utilization to better provision the service resources.
5. Compare changes in system behaviour after and before a new system configuration and/or new deployment.

(Taken from [2]).

## Black-Box Versus White-Box
* A black-box monitoring approach monitors a service from the user's perspective, it does not provide information about '_private_' internal state from a service, but more critical information like "_is the system actually working from a user's perspective?_".
* A white-box approach can collect internal information regarding the state of the system, e.g. current resource utilization like RAM or CPU. 

	"_We combine **heavy use of white-box monitoring** with **modest but critical uses of black-box monitoring**. The simplest way to think about black-box monitoring versus white-box monitoring is that black-box monitoring is **symptom-oriented** and represents **active** —not predicted— problems: "The system isn’t working correctly, right now." White-box monitoring depends on the ability to inspect the innards of the system, such as logs or HTTP endpoints, with instrumentation. White-box monitoring therefore allows detection of **imminent problems** [...]._" [1]

## The Four Golden Signals
1. **Latency**: Distinguish between the latency of packet delivery for errors (e.g. HTTP 500s, when a database in the backend is down) and the latency of packets without erros (HTTP 200s), since an HTTP error response can be very quick, and might skew the overall median latency towards a faster latency than what users are actually experiencing with correct responses.
2. **Traffic**: A measure of how much demand is being placed on the system. This signal should be a high-level system-specific metric, for example, the amount of HTTP requests per second. In the case of our implementation, a good metric might be the amount of active session per half hour or per hour, or even in a smaller time frame like 10min.
3. **Errors**: "_The rate of requests that fail, either explicitly (e.g., HTTP 500s), implicitly (for example, an HTTP 200 success response, but coupled with the wrong content), or by policy (for example, "If you committed to one-second response times, any request over one second is an error")._"[1]. In our system errors should be HTTP 500s responses, where something failed within the session management system. An HTTP 429 response ("Too many requests") should not be counted as an error.
4. **Saturation**: Measure system utilization, sometimes to proactively react to an incoming saturated system (e.g. a hard disk almost totally filled by database operations). "_Latency increases are often a leading indicator of saturation. Measuring your 99th percentile response time over some small window (e.g., one minute) can give a very early signal of saturation._" [1]. Both Docker and systems running in Go have some **hard limits** that should be tracked through metrics, in the case of Docker the amount of active private networks is critical and has a hard limit (see [Domain-specific monitoring](#domain-specific-monitoring-go-and-docker)).

### Application-specific signalling
|Signal | Black-box | White-box | Prometheus metrics | 
| --- | --- | --- | --- |
| Latency | Measure the total time to process a successful (200) HTTP request for a new session. Take into consideration that **timeouts** can take place and that error HTTP responses might have different latencies.| It could also be possible to explicitly monitor the amount of time needed to create a new session (within the session creation daemon), since a very big divergence from the mean could also be a clue about problems with the connection with the Docker daemon. **TODO**: timeout scd if it takes to long to create a new session (this happened previously when Docker ran out of IP space). | **Histogram**: measure in which range 99% of the requests' latency is. Outliers can make the user experience very bad. |
| Traffic | None | Amount of **different** HTTP requests per minute, decompose the traffic in successful requests for a session, and failed ones, like _Too many requests_ responses. Furthermore, measure the amount of active (and idle sessions) as a proxy for the traffic.  | **Counter**: ideal type to measure a rate, like number of requests. Use the _rate()_ function in PromQL to calculate a rate (change in time). |
| Errors | Count the amount of 500s HTTP responses. Define a maximum amount of time that is acceptable as a response time for a requests. If a response takes longer than that it should be labelled as an error and counted as one as well. | In the first implementation none, but it would make sense to be able to read from the logs the kinds of errors that take place. | **Counter**: see argumentation for '_Traffic_'. |
| Saturation | None | Monitor the **hard limits** of both Docker and Go. Use the [exporters and integrations](https://prometheus.io/docs/instrumenting/exporters/) for Prometheus, especially for Docker. The hard limits are especially critical for Docker, since, for example, if the maximum amount of private networks is exceeded, Docker will not be able to create any more sessions. Prometheus can also monitor some other Go-specific stats, like amount of goroutines (which is an indirect way of seeing the system utilization). | **Gauge**: the metric type counter does not make sense in this scenario, since a counter can only be increased. A gauge changes its value to precisely the value at a given moment, e.g. the current amount of virtual private networks in a Docker deployment can both go up or down. | 
| Versioning | Executable's version, command-line flags during execution and (eventually) version of config file should be tracked (see section [Versioning, rollouts and monitoring](#versioning-rollouts-and-monitoring). | - | **Gauge**: track versions with string values. |

### Mean latency
When measuring the mean latency, it should also be important to quantify the latency of responses in a histogram, since even if the mean is small, the slowest responses can still deviate a lot from the mean. Basically, also analyze the distribution of response times, to see if even the slowest responses are acceptable. In the case of our system, it would be interesting to track the time needed to create a new session, serve a session to a user request, etc. Since for example, a degraded response time while serving a session to a user heavily affects the overall user experience.

### Status of served traffic (HTTP)
It is advisable to monitor the different HTTP responses individually in the metrics. It is also advisable to monitor HTTP responses blocking a resource for a user (as in our case with the response for _Too many requests_) [2]. Only by individualizing the data of HTTP responses it is possible to establish actionable responses to different situations seen in changes in responses rates (e.g. changes in the amount of error responses, while not taking into consideration the error responses blocking the users per quota utilization _Too many requests_).

## Granularity of measurements
There has to be considerations towards how frequently a measurement will be logged or even performed.

"_On the other hand, for a web service targeting no more than 9 hours aggregate downtime per year (99.9% annual uptime), probing for a 200 (success) status more than once or twice a minute is probably unnecessarily frequent. Similarly, checking hard drive fullness for a service targeting 99.9% availability more than once every 1–2 minutes is probably unnecessary._" [1].

Collecting data with a very high frequency can be an expensive operation. Data can always be "_reduced_" by aggregating it, reducing its collection frequency or by reducing its distribution with a given frequency to histogramic buckets.

## Versioning, rollouts and monitoring
While monitoring a running production system (that might undertake a rollout) it makes sense to also continuously monitor a series of version- and binary-specific information:

* The version of the binary currently running in production.
* The command-line flags with which the program is currently being executed.
* If the configuration data (e.g. config files) is tracked with a version, monitor any changes of it.

Monitoring these values helps to pin-point any problems related to changing the version of a running system. And makes figuring out the version of the currently deployed systems easier than looking for it in a repository or in a CI/CD pipeline. [2]

### Monitoring system configuration
The configuration of the monitoring system should be integrated within the revision control system in which the system code lies, since it is easier to then track and rollback changes to the monitoring system's configuration in case it is needed. An additional benefit is that the configuration files can then be automatically linted and checked (see [promtool](https://prometheus.io/docs/prometheus/latest/configuration/unit_testing_rules/) for Prometheus) [2].

## Domain-specific monitoring (Go and Docker)
Depending on the language of the deployment being monitored, there are some internals that should be tracked for _saturation_. In the case of Go, the number of **active goroutines** should be tracked.

Docker also provides plenty of internal information that can be retrieved through the API of the Docker daemon. The current implementation of the system has for instance a **hard limit** on the amount of private virtual networks that it can create for different sessions. If the **hard limit** is achieved, then, the session manager daemon cannot create any new sessions.

## Prometheus
Prometheus defines different **targets** from which metrics are periodically retrieved. These targets can range from whole servers, database applications, single applications or services, etc. 

### Metrics
Metrics are stored in a human-readable text format and are described by two main entries **TYPE** and **HELP**.

* **HELP**: Describes what the metric is.
* **TYPE**: 
	1. **Counter**: How many times has a particular event happened? E.g. number of requests. The value of this metric can only increase.
	2. **Gauge**: Measures the current value of something. E.g. current RAM memory usage, amount of active goroutines, etc. In contrast to counters, gauges can both increase and decrease their values depending on the current value being measured.
	3. **Histogram**: Places a measurement in different buckets representing different values. E.g. the different latency times for different responses, in order to not only measure the mean latency value, but also the worst possible latency values and how often they happen. 

Metrics can be custom-made inside an application using a Prometheus client library. Those metrics will be exposed through an HTTP server, so that Prometheus _Data retrieval worker_ can periodically collect the metrics from the applications.

### Further information and guides
* [Instrumenting a Go application for Prometheus](https://prometheus.io/docs/guides/go-application/): A guide with code snippets on how to instrument a Go application.
