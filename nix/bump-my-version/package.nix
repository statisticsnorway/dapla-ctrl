{
  buildPythonApplication,
  lib,
  hatchling,
  click,
  pydantic-settings,
  pydantic,
  questionary,
  rich,
  rich-click,
  tomlkit,
  wcmatch,
  fetchPypi,
}:

buildPythonApplication rec {
  pname = "bump-my-version";
  version = "0.28.1";
  pyproject = true;

  src = fetchPypi {
    pname = "bump_my_version";
    inherit version;
    hash = "sha256-5gje9Rkbr1BbbN6IvWeaCpX8TP6s5CR622CsD4p+V+4=";
  };

  build-system = [hatchling];

  dependencies = [
    click
    pydantic-settings
    pydantic
    questionary
    rich
    rich-click
    tomlkit
    wcmatch
  ];

  meta = {
    description = "A small CLI tool for releasing software by updating version strings in source code";
    longDescription = ''
      A small command line tool to simplify releasing software by updating all version strings in your source code by the correct increment and optionally commit and tag the changes.
    '';
    homepage = "https://callowayproject.github.io/bump-my-version";
    license = with lib.licenses; [mit];
  };
}
