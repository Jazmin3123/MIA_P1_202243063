package utils

import "testing" // uso testing para validar alias aceptados por el parser

func TestParseCommandAceptaAliasDeScripts(t *testing.T) { // esta prueba valida variantes comunes del archivo de prueba
	cmd, err := ParseCommand(`Mkuser -usr="user1" -pass=clave -grp=root`)
	if err != nil {
		t.Fatalf("parse devolvio error: %v", err)
	}

	if cmd.Name != "mkusr" || cmd.Params["user"] != "user1" {
		t.Fatalf("alias mkuser/usr no se normalizo: %#v", cmd)
	}

	cmd, err = ParseCommand(`mkfile -path=/a.txt -s=75`)
	if err != nil {
		t.Fatalf("parse devolvio error: %v", err)
	}

	if cmd.Params["size"] != "75" {
		t.Fatalf("alias -s no se normalizo: %#v", cmd)
	}
}

func TestValidateCommandAceptaBmBloc(t *testing.T) { // esta prueba valida el alias bm_bloc del reporte de bitmap de bloques
	cmd, err := ParseCommand(`rep -id=631A -path=/tmp/bm.txt -name=bm_bloc`)
	if err != nil {
		t.Fatalf("parse devolvio error: %v", err)
	}

	if err := ValidateCommand(cmd); err != nil {
		t.Fatalf("validate devolvio error: %v", err)
	}
}
