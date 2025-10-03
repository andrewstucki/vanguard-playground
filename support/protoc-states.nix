{ buildGoModule, lib, fetchFromGitHub }:

buildGoModule rec {
  pname = "protoc-states";
  version = "db7ca4073759649219ad072239ed0d674928d639";

  src = fetchFromGitHub {
    owner = "andrewstucki";
    repo = pname;
    rev = "${version}";
    sha256 = "sha256-bCS3cdsDYMyol/DPNeG9tEOQ+bAhk9rK6Jjx3WUNnzw=";
  };

  vendorHash = "sha256-nrH8Of9DlprkNtXqt29x6AJh1BIZExtfo4P83+HqZi8=";

  ldflags = [
    "-s"
    "-w"
  ];

  doCheck = false;

  meta = with lib; {
    description = "Protobuf generator for state machines";
    homepage = "https://github.com/andrewstucki/protoc-states";
    license = licenses.mit;
  };
}
