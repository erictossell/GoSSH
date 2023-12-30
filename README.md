# GoSSH

A multi-host SSH command line tool written in Go.

### Usage

`GoSSH "command" "command2" "command3"`

### Installation

#### Go Installation from Source

Clone the repository and change directory:
```sh
git clone https://github.com/erictossell/GoSSH.git && cd GoSSH
```

Install with Go:
```go
go install
```

#### NixOS Flakes Installation

In `flake.nix` inputs add:

```nix
inputs = {
  GoSHH.url = "github:erictossell/GoSSH";
}; 
```

In `flake.nix` modules add:

```nix
modules = [
  ({ pkgs, GoSSH, ... }: 
  {
    environment.systemPackages = with pkgs; [
      GoSSH.packages.${system}.GoSSH
    ];
  })
];
```
**or** 

Imported as a `module.nix`:

```nix
{ pkgs, GoSSH, ... }: 
{
  environment.systemPackages = with pkgs; [
    GoSSH.packages.${system}.GoSSH
  ];
}
```

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
