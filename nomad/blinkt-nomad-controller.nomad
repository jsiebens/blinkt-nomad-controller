job "blinkt-nomad-controller" {

  datacenters = ["dc1"]

  type = "system"

  group "blinkt-nomad-controller" {
    count = 1

    task "blinkt-nomad-controller" {
      driver = "raw_exec"

      artifact {
        source = "https://github.com/jsiebens/blinkt-nomad-controller/releases/download/v0.2-alpha/blinkt-nomad-controller_arm"
      }

      config {
        command = "local/blinkt-nomad-controller_arm"
      }

      resources {
        cpu    = 100
        memory = 128
        network {
          mbits = 10
        }
      }
    }
  }

}