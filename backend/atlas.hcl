variable "db_url" {
  type    = string
  default = getenv("POSTGRES_CONNECTION_STRING")
}

env "local" {
  migration {
    dir = "file://migrations"
  }

  url = var.db_url

  dev = "docker://postgres/17/alpine"

  format {
    migrate {
      apply = "{{ range .Pending }}{{ printf \"Applied %s\\n\" .Name }}{{ end }}{{ if .Error }}{{ printf \"Error: %s\\n\" .Error }}{{ end }}"
    }
  }
}