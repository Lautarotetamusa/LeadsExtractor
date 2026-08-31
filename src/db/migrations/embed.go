// Package migrations embebe los archivos .sql de este directorio para que
// goose los pueda aplicar desde el binario compilado, sin depender de que
// el filesystem de despliegue tenga la carpeta db/ presente.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
