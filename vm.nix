{ ... }:
let
  sources = import ./nix/sources.nix;
  pkgs = import sources.nixpkgs { };
in
{
  nix.nixPath = [
    "nixpkgs=${pkgs.path}"
  ];
  nixos-shell.mounts = {
    mountHome = false;
    mountNixProfile = false;
    cache = "none"; # default is "loose"

    extraMounts = {
      "/localpv" = {
        target = ./.;
        cache = "none";
      };
    };
  };

  virtualisation = {
    cores = 4;
    memorySize = 2048;
    # Uncomment to be able to ssh into the vm, example:
    # ssh -p 2222 -o StrictHostKeychecking=no root@localhost
    # forwardPorts = [
    #  { from = "host"; host.port = 2222; guest.port = 22; }
    # ];
    diskSize = 20 * 1024;
    docker = {
      enable = true;
    };
  };
  documentation.enable = false;

  networking = {
    firewall = {
      allowedTCPPorts = [
        6443 # k3s: required so that pods can reach the API server (running on port 6443 by default)
      ];
    };
  };

  services = {
    openssh.enable = true;
    k3s = {
      enable = true;
      role = "server";
      extraFlags = toString [
        "--disable=traefik"
      ];
    };
  };

  programs.git = {
    enable = true;
    config = {
      safe = {
        directory = [ "/localpv" ];
      };
    };
  };

  systemd.tmpfiles.rules = [
    "L+ /usr/local/bin - - - - /run/current-system/sw/bin/"
  ];

  environment = {
    variables = {
      KUBECONFIG = "/etc/rancher/k3s/k3s.yaml";
      CI_K3S = "true";
      GOPATH = "/localpv/nix/.go";
      EDITOR = "vim";
    };

    shellAliases = {
      k = "kubectl";
      ke = "kubectl -n openebs";
    };

    shellInit = ''
      export PATH=$GOPATH/bin:$PATH
      cd /localpv
    '';

    systemPackages = with pkgs; [ vim docker-client k9s e2fsprogs xfsprogs ];
  };
}
