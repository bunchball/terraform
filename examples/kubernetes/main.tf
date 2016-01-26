provider "kubernetes" {
  endpoint = "http://<yourHostHere>:8080"
  username = "blah"
  password = "blah"
  insecure = true
  client_certificate = ""
  client_key = ""
  cluster_ca_certificate = ""
}

resource "kubernetes_namespace" "ns" {
  name = "example"
  labels {
    name = "iAmANamespace"
  }
}

resource "kubernetes_pod" "pod" {
  name = "i-am-a-pod"
  namespace = "${kubernetes_namespace.ns.name}"
  labels {
    app = "demo"
    version = "1.0.0"
    track = "prod"
    topic = "pod"
  }
  container {
    name = "helloworld"
    image = "b.gcr.io/kuar/helloworld:1.0.0"
    port {
      name = "one"
      protocol = "TCP"
      containerPort = "80"
    }
    port {
      name = "two"
      protocol = "TCP"
      containerPort = "81"
    }
  }
  container {
    name = "helloworld2"
    image = "gcr.io/google_containers/pause"
  }
  terminationGracePeriodSeconds = "30"
}

resource "kubernetes_replication_controller" "rc" {
  name = "rc"
  namespace = "${kubernetes_namespace.ns.name}"
  replicas = "5"
  selector {
    app = "demo"
    version = "1.0.0"
    track = "prod"
    topic = "rc"
  }
  labels {
    app = "demo"
    topic = "rc"
    nl1 = "new_label"
    topic = "rc"
  }
  pod {
    namespace = "${kubernetes_namespace.ns.name}" #have to specify namespace here because of a bug
    labels {
      app = "demo"
      version = "1.0.0"
      track = "prod"
      topic = "rc"
    }
    container {
      name = "helloworld"
      image = "b.gcr.io/kuar/helloworld:1.0.0"
      port {
        name = "one"
        protocol = "TCP"
        containerPort = "80"
      }
      port {
        name = "two"
        protocol = "TCP"
        containerPort = "81"
      }
    }
    container {
      name = "helloworld2"
      image = "gcr.io/google_containers/pause"
    }
    terminationGracePeriodSeconds = "30"
  }
}

resource "kubernetes_service" "svc" {
  name = "svc"
  namespace = "${kubernetes_namespace.ns.name}"
  selector {
    app = "demo"
  }
  labels {
    app = "demo"
    topic = "service"
  }
  port {
    name = "one"
    protocol = "TCP"
    port = "80"
    targetPort = "one"
  }
}
