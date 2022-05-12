{ pkgs
, config
, astra ? (import ../. { inherit pkgs; })

}: rec {
  start-astra = pkgs.writeShellScriptBin "start-astra" ''
    # rely on environment to provide astrad
    export PATH=${pkgs.pystarport}/bin:$PATH
    ${../scripts/start-astra} ${config.astra-config} ${config.dotenv} $@
  '';
  start-scripts = pkgs.symlinkJoin {
    name = "start-scripts";
    paths = [ start-astra ];
  };
}