// used to name infrastructure (CHANGE THIS)
name = "smoke-test"

launchpad = {
  drain = false

  mcr_version = "23.0.7"
  mke_version = "3.7.3"
  msr_version = ""

  mke_connect = {
    username = "admin"
    password = "m!rantis2024"
    insecure = false
  }
}

// configure the network stack
network = {
  cidr                 = "172.31.0.0/16"
  public_subnet_count  = 3
  private_subnet_count = 0 // if 0 then no private nodegroups allowed
}

// one definition for each group of machines to include in the stack
nodegroups = {
  "ACon" = { // managers for A group
    platform    = "ubuntu_22.04"
    count       = 1
    type        = "m6a.2xlarge"
    volume_size = 100
    role        = "manager"
    public      = true
    user_data   = ""
  },
  "AWrk_Ubu22" = { // workers for A group
    platform    = "ubuntu_22.04"
    count       = 1
    type        = "c6a.xlarge"
    volume_size = 100
    public      = true
    role        = "worker"
    user_data   = ""
  },
  "AWrk_Windows2022" = {
    "platform" : "windows_core_2022",
    "count" : 1,
    "type" : "c6a.xlarge",
    "volume_size" : "100",
    "role" : "worker",
    "public" : true,
    "user_data" : "",
  },
}

// set a windows password, if you have windows nodes
# windows_password = "testp@ss!"
