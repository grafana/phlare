---
title: "Getting started"
menuTitle: "Getting started"
description: "Getting started sending profiles with Phlare"
weight: 10
---

# Getting started

Phlare is a continuous profiling database that allows you to analyze the performance of your applications. When sending profiles to Phlare, you have two options: using the Grafana agent in pull mode or utilizing the Phlare SDKs in push mode. This document will provide an overview of these two methods and guide you on when to choose each option.

![phlare_agent_server_diag](https://github.com/grafana/phlare/assets/23323466/9b7e2255-d54f-4e51-b81b-d98baab904e6)

## Grafana Agent (Pull Mode)

The Grafana agent is a component that runs alongside your application and periodically pulls the profiles from it. This mode is suitable when you want to collect profiles from existing applications without modifying their source code. Here's how it works:

1. Install and configure the Grafana agent on the same machine or container where your application is running.
2. The agent will periodically query your application's performance profiling endpoints, such as pprof endpoints in Go applications.
3. The retrieved profiles are then sent to the Phlare server for storage and analysis.

Using the Grafana agent is a convenient option when you have multiple applications or microservices, as you can centralize the profiling process without making any changes to your application's codebase.

## Phlare SDKs (Push Mode)

Alternatively, you can use the Phlare SDKs to push profiles from your application directly to the Phlare server. This mode is suitable when you want to have more control over the profiling process or when the application you are profiling is written in a language supported by the SDKs (e.g., Ruby, Python, etc.). Follow these steps to use the Phlare SDKs:

1. Install the relevant Phlare SDK for your application's programming language (e.g., Ruby gem, pip package, etc.)
2. Instrument your application's code using the SDK to capture the necessary profiling data
3. Periodically push the captured profiles to the Phlare server for storage and analysis

By using the Phlare SDKs, you have the flexibility to customize the profiling process according to your application's specific requirements. You can selectively profile specific sections of code or send profiles at different intervals, depending on your needs.

## Choosing the Right Mode

The decision of which mode to use depends on your specific use case and requirements. Here are some factors to consider when making the choice:

- Ease of setup: If you want a quick and straightforward setup without modifying your application's code, the Grafana agent in pull mode is a good choice
- Language support: If your application is written in a language supported by the Phlare SDKs and you want more control over the profiling process, using the SDKs in push mode is recommended
- Flexibility: The Phlare SDKs provide more flexibility in terms of customizing the profiling process and capturing specific sections of code with labels. If you have specific profiling needs or want to fine-tune the data collection process, the SDKs offer greater flexibility

If you have more questions feel free to reach out in our Slack channel or create an issue on github and the Phlare team will 