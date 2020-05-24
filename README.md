# A Nomad Controller for the Pimoroni Blinkt on a Raspberry Pi #

A simple way to physically/visually display the number of allocations running on Raspberry Pi-based [HashiCorp](https://www.hashicorp.com) [Nomad](https://github.com/hashicorp/nomad) worker nodes by using a [Pimoroni Blinkt!](https://shop.pimoroni.com/products/blinkt).
Alternatively, it can display the amount of allocated cpu, memory, disk or network of the worker node.

The Blinkt is a low-profile strip of eight super-bright, color LED indicators that plugs directly onto the Raspberry Pi's GPIO header. Several available software libraries make it easy to control the color and brightness of each LED independently.

## How It Works ##

This little tool is designed to be deployed as a Nomad [system](https://www.nomadproject.io/docs/schedulers/#system) job
Once deployed, every running allocation that lands on a node will cause an LED indicator on that node's Blinkt to turn on. (only the first 8 Pods can be displayed). 
As new jobs and allocations get created or deleted the light display will adjust accordingly.

The controller will scrape the metrics of the node it is running an via the [metrics HTTP api](https://www.nomadproject.io/api-docs/metrics/).

## Acknowledgements ##

This project is based on the [blinkt-k8s-controller](https://github.com/apprenda/blinkt-k8s-controller) of @apprenda
and draws inspiration and borrows heavily from the work done by @alexellis on [Docker on Raspberry Pis](http://blog.alexellis.io/visiting-pimoroni/) and his [Blinkt Go libraries](https://github.com/alexellis/blinkt_go), themselves based on work by @gamaral for using the `/sys/` fs interface [instead of special libraries or elevated privileges](https://guillermoamaral.com/read/rpi-gpio-c-sysfs/) to `/dev/mem` on the Raspberry Pi.

## Requirements ##

A Raspberry Pi-based Nomad cluster, where the raw_exec driver is enabled on the worker nodes.

Physically install a [Pimoroni Blinkt](https://shop.pimoroni.com/products/blinkt) on all the Raspberry Pi worker nodes you want to use for display. **No additional sofware or setup is required for the Blinkt**.

## Usage ##

Plan and run the Nomad job using the included job definition:

```sh
nomad plan nomad/blinkt-nomad-controller.nomad
```

```sh
nomad run nomad/blinkt-nomad-controller.nomad
```

If you want to monitor and display other resources, like allocated memory, adjust the command accordingly:

```sh
blinkt-nomad-controller_arm -resource=memory
```