# GoSSH

A multi-host SSH command line tool written in Go.

### Usage

`GoSSH "command" "command2" "command3"`


#### Configuration

The first time running the application will generate a `.config/GoSSH/configuration.json` if one does not exist.

##### Example Configuration

```json
{
  "servers": ["192.168.2.195", "192.168.2.196", "192.168.2.197"],
  "ssh_options": {
    "192.168.2.195": "-p 2973",
    "192.168.2.196": "-p 2973",
    "192.168.2.197": "-p 2973"
  },
  "users": {
    "192.168.2.195": "eriim",
    "192.168.2.196": "eriim",
    "192.168.2.197": "eriim"
  }
}
```
