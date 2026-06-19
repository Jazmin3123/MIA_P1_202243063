package comandos

import (
	"fmt"           // uso fmt para armar textos de reportes
	"html"          // uso html para escapar textos dentro de graphviz
	"os"            // uso os para crear carpetas y archivos
	"os/exec"       // uso exec para llamar graphviz cuando sea posible
	"path/filepath" // uso filepath para crear carpetas y detectar extensiones
	"strings"       // uso strings para construir dot y normalizar nombres
	"time"          // uso time para formatear fechas en reportes

	"MIA_P1_202243063/estructuras" // uso estructuras para leer datos binarios
	"MIA_P1_202243063/utils"       // uso utils para leer structs y limpiar bytes
)

func EjecutarREP(params map[string]string) error { // esta funcion ejecuta el comando rep
	mounted, exists := FindMountedPartitionByID(params["id"]) // busco la particion montada del reporte
	if !exists {
		return fmt.Errorf("no existe una particion montada con id %s", params["id"])
	}

	name := strings.ToLower(params["name"])
	var err error
	switch name {
	case "mbr":
		err = reporteMBR(*mounted, params["path"])
	case "disk":
		err = reporteDISK(*mounted, params["path"])
	case "sb":
		err = reporteSB(*mounted, params["path"])
	case "bm_inode":
		err = reporteBitmap(*mounted, params["path"], true)
	case "bm_block", "bm_bloc":
		err = reporteBitmap(*mounted, params["path"], false)
	case "file":
		err = reporteFILE(*mounted, params["path"], params["path_file_ls"])
	case "inode":
		err = reporteINODE(*mounted, params["path"])
	case "block":
		err = reporteBLOCK(*mounted, params["path"])
	case "ls":
		err = reporteLS(*mounted, params["path"], params["path_file_ls"])
	case "tree":
		err = reporteTREE(*mounted, params["path"])
	default:
		return fmt.Errorf("no se pudo generar reporte %s: reporte no implementado", name)
	}

	if err != nil {
		return fmt.Errorf("no se pudo generar reporte %s: %w", name, err)
	}

	return nil
}

func reporteMBR(mounted MountedPartition, outputPath string) error { // esta funcion genera el reporte mbr
	file, err := os.Open(mounted.Path)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	var mbr estructuras.MBR
	if err := utils.ReadStructAt(file, 0, &mbr); err != nil {
		return err
	}

	var dot strings.Builder
	dot.WriteString("digraph G {\n")
	dot.WriteString("node [shape=plaintext];\n")
	dot.WriteString("mbr [label=<\n")
	dot.WriteString("<table border='1' cellborder='1' cellspacing='0'>\n")
	dot.WriteString("<tr><td colspan='2'><b>MBR</b></td></tr>\n")
	dot.WriteString(fmt.Sprintf("<tr><td>mbr_tamano</td><td>%d</td></tr>\n", mbr.Size))
	dot.WriteString(fmt.Sprintf("<tr><td>mbr_fecha_creacion</td><td>%d</td></tr>\n", mbr.CreationDate))
	dot.WriteString(fmt.Sprintf("<tr><td>mbr_dsk_signature</td><td>%d</td></tr>\n", mbr.DiskSignature))
	dot.WriteString(fmt.Sprintf("<tr><td>dsk_fit</td><td>%c</td></tr>\n", mbr.Fit))

	for index, partition := range mbr.Partitions {
		if partition.Size == 0 {
			continue
		}

		dot.WriteString(fmt.Sprintf("<tr><td colspan='2'><b>partition_%d</b></td></tr>\n", index+1))
		dot.WriteString(fmt.Sprintf("<tr><td>part_status</td><td>%d</td></tr>\n", partition.Status))
		dot.WriteString(fmt.Sprintf("<tr><td>part_type</td><td>%c</td></tr>\n", partition.Type))
		dot.WriteString(fmt.Sprintf("<tr><td>part_fit</td><td>%c</td></tr>\n", partition.Fit))
		dot.WriteString(fmt.Sprintf("<tr><td>part_start</td><td>%d</td></tr>\n", partition.Start))
		dot.WriteString(fmt.Sprintf("<tr><td>part_size</td><td>%d</td></tr>\n", partition.Size))
		dot.WriteString(fmt.Sprintf("<tr><td>part_name</td><td>%s</td></tr>\n", utils.BytesToString(partition.Name[:])))

		if partition.Type == 'E' {
			if err := escribirEBRsEnReporteMBR(file, &dot, partition); err != nil {
				return err
			}
		}
	}

	dot.WriteString("</table>>];\n")
	dot.WriteString("}\n")

	return escribirGraphviz(outputPath, dot.String())
}

func escribirEBRsEnReporteMBR(file *os.File, dot *strings.Builder, extended estructuras.Partition) error { // esta funcion agrega ebr al reporte mbr cuando hay particiones logicas
	currentPosition := extended.Start

	for currentPosition != -1 {
		var ebr estructuras.EBR
		if err := utils.ReadStructAt(file, int64(currentPosition), &ebr); err != nil {
			return err
		}

		if ebr.Size == 0 {
			break
		}

		dot.WriteString("<tr><td colspan='2'><b>EBR</b></td></tr>\n")
		dot.WriteString(fmt.Sprintf("<tr><td>part_mount</td><td>%d</td></tr>\n", ebr.Mount))
		dot.WriteString(fmt.Sprintf("<tr><td>part_fit</td><td>%c</td></tr>\n", ebr.Fit))
		dot.WriteString(fmt.Sprintf("<tr><td>part_start</td><td>%d</td></tr>\n", ebr.Start))
		dot.WriteString(fmt.Sprintf("<tr><td>part_size</td><td>%d</td></tr>\n", ebr.Size))
		dot.WriteString(fmt.Sprintf("<tr><td>part_next</td><td>%d</td></tr>\n", ebr.Next))
		dot.WriteString(fmt.Sprintf("<tr><td>part_name</td><td>%s</td></tr>\n", utils.BytesToString(ebr.Name[:])))

		currentPosition = ebr.Next
	}

	return nil
}

func reporteDISK(mounted MountedPartition, outputPath string) error { // esta funcion genera el reporte visual de uso del disco
	file, err := os.Open(mounted.Path)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	var mbr estructuras.MBR
	if err := utils.ReadStructAt(file, 0, &mbr); err != nil {
		return err
	}

	used := usedPrimaryPartitions(mbr)
	var dot strings.Builder
	dot.WriteString("digraph G {\n")
	dot.WriteString("node [shape=plaintext];\n")
	dot.WriteString("disk [label=<\n")
	dot.WriteString("<table border='1' cellborder='1' cellspacing='0'><tr>\n")
	dot.WriteString(fmt.Sprintf("<td>MBR<br/>%.2f%%</td>\n", porcentaje(float64(utils.StructSize(estructuras.MBR{})), float64(mbr.Size))))

	current := utils.StructSize(estructuras.MBR{})
	for _, partition := range used {
		if partition.Start > current {
			free := partition.Start - current
			dot.WriteString(fmt.Sprintf("<td>Libre<br/>%.2f%%</td>\n", porcentaje(float64(free), float64(mbr.Size))))
		}

		if partition.Type == 'E' {
			dot.WriteString("<td>\n")
			dot.WriteString(fmt.Sprintf("<table border='1' cellborder='1' cellspacing='0'><tr><td colspan='10'>Extendida<br/>%.2f%%</td></tr><tr>\n", porcentaje(float64(partition.Size), float64(mbr.Size))))
			if err := escribirParticionesLogicasDISK(file, &dot, partition); err != nil {
				return err
			}
			dot.WriteString("</tr></table>\n</td>\n")
		} else {
			dot.WriteString(fmt.Sprintf("<td>%s<br/>%c<br/>%.2f%%</td>\n", html.EscapeString(utils.BytesToString(partition.Name[:])), partition.Type, porcentaje(float64(partition.Size), float64(mbr.Size))))
		}
		current = partition.Start + partition.Size
	}

	if current < mbr.Size {
		free := mbr.Size - current
		dot.WriteString(fmt.Sprintf("<td>Libre<br/>%.2f%%</td>\n", porcentaje(float64(free), float64(mbr.Size))))
	}

	dot.WriteString("</tr></table>>];\n")
	dot.WriteString("}\n")

	return escribirGraphviz(outputPath, dot.String())
}

func escribirParticionesLogicasDISK(file *os.File, dot *strings.Builder, extended estructuras.Partition) error { // esta funcion escribe ebr y logicas dentro del reporte disk
	ebrSize := utils.StructSize(estructuras.EBR{})
	currentPosition := extended.Start
	currentEnd := extended.Start

	for currentPosition != -1 {
		var ebr estructuras.EBR
		if err := utils.ReadStructAt(file, int64(currentPosition), &ebr); err != nil {
			return err
		}

		if ebr.Size == 0 {
			dot.WriteString("<td>EBR</td>")
			currentEnd = currentPosition + ebrSize
			break
		}

		if ebr.Start > currentEnd {
			free := ebr.Start - currentEnd
			dot.WriteString(fmt.Sprintf("<td>Libre<br/>%.2f%%</td>", porcentaje(float64(free), float64(extended.Size))))
		}

		dot.WriteString("<td>EBR</td>")
		dot.WriteString(fmt.Sprintf("<td>Logica<br/>%s<br/>%.2f%%</td>", html.EscapeString(utils.BytesToString(ebr.Name[:])), porcentaje(float64(ebr.Size), float64(extended.Size))))
		currentEnd = ebr.Start + ebrSize + ebr.Size
		currentPosition = ebr.Next
	}

	extendedEnd := extended.Start + extended.Size
	if currentEnd < extendedEnd {
		free := extendedEnd - currentEnd
		dot.WriteString(fmt.Sprintf("<td>Libre<br/>%.2f%%</td>", porcentaje(float64(free), float64(extended.Size))))
	}

	return nil
}

func reporteSB(mounted MountedPartition, outputPath string) error { // esta funcion genera el reporte del superbloque
	file, sb, err := abrirSuperBloque(mounted)
	if err != nil {
		return err
	}
	defer file.Close()

	dot := fmt.Sprintf(`digraph G {
node [shape=plaintext];
sb [label=<
<table border='1' cellborder='1' cellspacing='0'>
<tr><td colspan='2'><b>SUPER BLOQUE</b></td></tr>
<tr><td>s_filesystem_type</td><td>%d</td></tr>
<tr><td>s_inodes_count</td><td>%d</td></tr>
<tr><td>s_blocks_count</td><td>%d</td></tr>
<tr><td>s_free_blocks_count</td><td>%d</td></tr>
<tr><td>s_free_inodes_count</td><td>%d</td></tr>
<tr><td>s_mtime</td><td>%d</td></tr>
<tr><td>s_umtime</td><td>%d</td></tr>
<tr><td>s_mnt_count</td><td>%d</td></tr>
<tr><td>s_magic</td><td>%x</td></tr>
<tr><td>s_inode_s</td><td>%d</td></tr>
<tr><td>s_block_s</td><td>%d</td></tr>
<tr><td>s_first_ino</td><td>%d</td></tr>
<tr><td>s_first_blo</td><td>%d</td></tr>
<tr><td>s_bm_inode_start</td><td>%d</td></tr>
<tr><td>s_bm_block_start</td><td>%d</td></tr>
<tr><td>s_inode_start</td><td>%d</td></tr>
<tr><td>s_block_start</td><td>%d</td></tr>
</table>>];
}
`, sb.FilesystemType, sb.InodesCount, sb.BlocksCount, sb.FreeBlocksCount, sb.FreeInodesCount, sb.MTime, sb.UMTime, sb.MntCount, sb.Magic, sb.InodeSize, sb.BlockSize, sb.FirstInode, sb.FirstBlock, sb.BmInodeStart, sb.BmBlockStart, sb.InodeStart, sb.BlockStart)

	return escribirGraphviz(outputPath, dot)
}

func reporteBitmap(mounted MountedPartition, outputPath string, inodeBitmap bool) error { // esta funcion genera bm_inode o bm_block en texto
	file, sb, err := abrirSuperBloque(mounted)
	if err != nil {
		return err
	}
	defer file.Close()

	count := sb.BlocksCount
	start := sb.BmBlockStart
	if inodeBitmap {
		count = sb.InodesCount
		start = sb.BmInodeStart
	}

	bitmap := make([]byte, count)
	if _, err := file.ReadAt(bitmap, int64(start)); err != nil {
		return fmt.Errorf("no se pudo leer bitmap: %w", err)
	}

	var text strings.Builder
	for index, value := range bitmap {
		if value == 0 {
			value = '0'
		}
		text.WriteByte(value)
		if (index+1)%20 == 0 {
			text.WriteByte('\n')
		} else {
			text.WriteByte(' ')
		}
	}

	text.WriteByte('\n')
	return escribirArchivoReporte(outputPath, text.String())
}

func reporteFILE(mounted MountedPartition, outputPath string, filePath string) error { // esta funcion genera el reporte file en texto
	file, sb, err := abrirSuperBloque(mounted)
	if err != nil {
		return err
	}
	defer file.Close()

	contenido, err := leerArchivoParaReporte(file, sb, filePath, mounted)
	if err != nil {
		return err
	}

	return escribirArchivoReporte(outputPath, fmt.Sprintf("archivo: %s\n%s\n", filePath, contenido))
}

func reporteINODE(mounted MountedPartition, outputPath string) error { // esta funcion genera el reporte de inodos usados
	file, sb, err := abrirSuperBloque(mounted)
	if err != nil {
		return err
	}
	defer file.Close()

	used, err := inodosUsados(file, sb)
	if err != nil {
		return err
	}

	var dot strings.Builder
	dot.WriteString("digraph G {\nnode [shape=plaintext];\n")

	for _, item := range used {
		inode := item.Inode
		dot.WriteString(fmt.Sprintf("inode%d [label=<\n", item.Index))
		dot.WriteString("<table border='1' cellborder='1' cellspacing='0'>\n")
		dot.WriteString(fmt.Sprintf("<tr><td colspan='2'><b>INODO %d</b></td></tr>\n", item.Index))
		dot.WriteString(fmt.Sprintf("<tr><td>i_uid</td><td>%d</td></tr>\n", inode.UID))
		dot.WriteString(fmt.Sprintf("<tr><td>i_gid</td><td>%d</td></tr>\n", inode.GID))
		dot.WriteString(fmt.Sprintf("<tr><td>i_s</td><td>%d</td></tr>\n", inode.Size))
		dot.WriteString(fmt.Sprintf("<tr><td>i_atime</td><td>%d</td></tr>\n", inode.ATime))
		dot.WriteString(fmt.Sprintf("<tr><td>i_ctime</td><td>%d</td></tr>\n", inode.CTime))
		dot.WriteString(fmt.Sprintf("<tr><td>i_mtime</td><td>%d</td></tr>\n", inode.MTime))
		for index, pointer := range inode.Block {
			dot.WriteString(fmt.Sprintf("<tr><td>i_block_%d</td><td>%d</td></tr>\n", index, pointer))
		}
		dot.WriteString(fmt.Sprintf("<tr><td>i_type</td><td>%d</td></tr>\n", inode.Type))
		dot.WriteString(fmt.Sprintf("<tr><td>i_perm</td><td>%s</td></tr>\n", utils.BytesToString(inode.Perm[:])))
		dot.WriteString("</table>>];\n")
	}

	dot.WriteString("}\n")
	return escribirGraphviz(outputPath, dot.String())
}

func reporteBLOCK(mounted MountedPartition, outputPath string) error { // esta funcion genera el reporte de bloques usados
	file, sb, err := abrirSuperBloque(mounted)
	if err != nil {
		return err
	}
	defer file.Close()

	used, err := inodosUsados(file, sb)
	if err != nil {
		return err
	}

	var dot strings.Builder
	dot.WriteString("digraph G {\nnode [shape=plaintext];\n")
	visitados := map[int32]bool{}

	for _, item := range used {
		for _, pointer := range item.Inode.Block {
			if pointer == -1 || visitados[pointer] {
				continue
			}
			visitados[pointer] = true

			if item.Inode.Type == estructuras.FolderBlockType {
				block, err := leerBloqueCarpeta(file, sb, pointer)
				if err != nil {
					return err
				}
				escribirNodoBloqueCarpeta(&dot, pointer, block)
			} else {
				var block estructuras.FileBlock
				if err := utils.ReadStructAt(file, int64(sb.BlockStart+pointer*sb.BlockSize), &block); err != nil {
					return err
				}
				escribirNodoBloqueArchivo(&dot, pointer, block)
			}
		}
	}

	dot.WriteString("}\n")
	return escribirGraphviz(outputPath, dot.String())
}

func reporteLS(mounted MountedPartition, outputPath string, targetPath string) error { // esta funcion genera reporte ls de una carpeta
	file, sb, err := abrirSuperBloque(mounted)
	if err != nil {
		return err
	}
	defer file.Close()

	inodeIndex, err := buscarInodoParaReporte(file, sb, targetPath, mounted)
	if err != nil {
		return err
	}

	inode, err := leerInodoPorIndice(file, sb, inodeIndex)
	if err != nil {
		return err
	}

	if inode.Type != estructuras.FolderBlockType {
		return fmt.Errorf("ls requiere una ruta de carpeta")
	}

	if sesion, activa := sesionParaReporte(mounted); activa {
		if err := validarPermisoInodo(inode, sesion.Usuario, permisoLectura, "listar la carpeta"); err != nil {
			return err
		}
	}

	var dot strings.Builder
	dot.WriteString("digraph G {\nnode [shape=plaintext];\n")
	dot.WriteString("ls [label=<\n<table border='1' cellborder='1' cellspacing='0'>\n")
	dot.WriteString("<tr><td>Permisos</td><td>Owner</td><td>Grupo</td><td>Size</td><td>Fecha Creacion</td><td>Fecha Modificacion</td><td>Tipo</td><td>Name</td></tr>\n")

	entradas, err := entradasCarpeta(file, sb, inode)
	if err != nil {
		return err
	}

	usuariosPorID, gruposPorID := mapasUsuariosParaReporte(mounted)

	for _, entry := range entradas {
		if entry.Nombre == "." || entry.Nombre == ".." {
			continue
		}

		child, err := leerInodoPorIndice(file, sb, entry.Inode)
		if err != nil {
			return err
		}

		tipo := "archivo"
		if child.Type == estructuras.FolderBlockType {
			tipo = "carpeta"
		}

		owner := nombreUsuarioReporte(usuariosPorID, child.UID)
		grupo := nombreGrupoReporte(gruposPorID, child.GID)
		dot.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
			permisosLinux(child), html.EscapeString(owner), html.EscapeString(grupo), child.Size, formatoFechaReporte(child.CTime), formatoFechaReporte(child.MTime), tipo, html.EscapeString(entry.Nombre)))
	}

	dot.WriteString("</table>>];\n}\n")
	return escribirGraphviz(outputPath, dot.String())
}

func leerArchivoParaReporte(file *os.File, sb estructuras.SuperBlock, filePath string, mounted MountedPartition) (string, error) { // esta funcion lee archivos para rep file respetando sesion si existe
	if sesion, activa := sesionParaReporte(mounted); activa {
		return leerArchivoPorRutaConUsuario(file, sb, filePath, sesion.Usuario)
	}

	return leerArchivoPorRuta(file, sb, filePath)
}

func buscarInodoParaReporte(file *os.File, sb estructuras.SuperBlock, path string, mounted MountedPartition) (int32, error) { // esta funcion busca rutas para reportes respetando sesion si existe
	if sesion, activa := sesionParaReporte(mounted); activa {
		return buscarInodoPorRutaConUsuario(file, sb, path, sesion.Usuario)
	}

	return buscarInodoPorRuta(file, sb, path)
}

func sesionParaReporte(mounted MountedPartition) (SesionSistema, bool) { // esta funcion usa permisos en reportes cuando la sesion activa pertenece a la misma particion
	if !sesionActual.Activa {
		return SesionSistema{}, false
	}

	if sesionActual.Montada.ID != mounted.ID {
		return SesionSistema{}, false
	}

	return sesionActual, true
}

func mapasUsuariosParaReporte(mounted MountedPartition) (map[int32]string, map[int32]string) { // esta funcion crea mapas uid/gid para mostrar nombres en ls
	usuariosPorID := map[int32]string{}
	gruposPorID := map[int32]string{}
	usuarios, grupos, err := leerUsuariosYGrupos(mounted)
	if err != nil {
		return usuariosPorID, gruposPorID
	}

	for _, usuario := range usuarios {
		if usuario.Activo {
			usuariosPorID[usuario.UID] = usuario.Usuario
		}
	}

	for _, grupo := range grupos {
		if grupo.Activo {
			gruposPorID[grupo.GID] = grupo.Nombre
		}
	}

	return usuariosPorID, gruposPorID
}

func nombreUsuarioReporte(usuarios map[int32]string, uid int32) string { // esta funcion devuelve nombre de usuario o uid si no se encuentra
	if nombre, exists := usuarios[uid]; exists {
		return nombre
	}

	return fmt.Sprintf("%d", uid)
}

func nombreGrupoReporte(grupos map[int32]string, gid int32) string { // esta funcion devuelve nombre de grupo o gid si no se encuentra
	if nombre, exists := grupos[gid]; exists {
		return nombre
	}

	return fmt.Sprintf("%d", gid)
}

func formatoFechaReporte(timestamp int64) string { // esta funcion formatea fechas unix para que el reporte sea legible
	if timestamp == 0 {
		return "-"
	}

	return time.Unix(timestamp, 0).Format("02/01/2006 15:04")
}

func permisosLinux(inode estructuras.Inode) string { // esta funcion convierte permisos octales a texto estilo linux
	prefix := "-"
	if inode.Type == estructuras.FolderBlockType {
		prefix = "d"
	}

	permisos := utils.BytesToString(inode.Perm[:])
	if len(permisos) != 3 {
		return prefix + "---------"
	}

	var salida strings.Builder
	salida.WriteString(prefix)
	for _, digit := range permisos {
		value := int(digit - '0')
		if value&4 != 0 {
			salida.WriteByte('r')
		} else {
			salida.WriteByte('-')
		}

		if value&2 != 0 {
			salida.WriteByte('w')
		} else {
			salida.WriteByte('-')
		}

		if value&1 != 0 {
			salida.WriteByte('x')
		} else {
			salida.WriteByte('-')
		}
	}

	return salida.String()
}

func reporteTREE(mounted MountedPartition, outputPath string) error { // esta funcion genera un arbol basico de inodos y bloques desde la raiz
	file, sb, err := abrirSuperBloque(mounted)
	if err != nil {
		return err
	}
	defer file.Close()

	var dot strings.Builder
	dot.WriteString("digraph G {\nrankdir=LR;\nnode [shape=plaintext];\n")
	visitados := map[int32]bool{}

	if err := escribirTreeInodo(file, sb, 0, &dot, visitados); err != nil {
		return err
	}

	dot.WriteString("}\n")
	return escribirGraphviz(outputPath, dot.String())
}

func abrirSuperBloque(mounted MountedPartition) (*os.File, estructuras.SuperBlock, error) { // esta funcion abre el disco y lee el superbloque
	file, err := os.Open(mounted.Path)
	if err != nil {
		return nil, estructuras.SuperBlock{}, fmt.Errorf("no se pudo abrir el disco: %w", err)
	}

	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(mounted.Partition.Start), &sb); err != nil {
		file.Close()
		return nil, estructuras.SuperBlock{}, err
	}

	if sb.Magic != estructuras.Ext2Magic {
		file.Close()
		return nil, estructuras.SuperBlock{}, fmt.Errorf("la particion no esta formateada como ext2")
	}

	return file, sb, nil
}

func escribirGraphviz(outputPath string, dot string) error { // esta funcion escribe dot e intenta renderizar con graphviz
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("no se pudo crear carpeta del reporte: %w", err)
	}

	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(outputPath)), ".")
	if extension == "" || extension == "dot" {
		if err := os.WriteFile(outputPath, []byte(dot), 0644); err != nil {
			return fmt.Errorf("no se pudo escribir reporte dot: %w", err)
		}
		fmt.Printf("reporte generado: %s\n", outputPath)
		return nil
	}

	dotPath := outputPath + ".dot"
	if err := os.WriteFile(dotPath, []byte(dot), 0644); err != nil {
		return fmt.Errorf("no se pudo escribir dot: %w", err)
	}

	cmd := exec.Command("dot", "-T"+extension, dotPath, "-o", outputPath)
	salida, err := cmd.CombinedOutput()
	if err != nil {
		detalle := strings.TrimSpace(string(salida))
		if detalle == "" {
			detalle = err.Error()
		}
		return fmt.Errorf("graphviz fallo para %s: %s; dot guardado en %s", outputPath, detalle, dotPath)
	}

	fmt.Printf("reporte generado: %s\n", outputPath)
	return nil
}

func escribirArchivoReporte(outputPath string, content string) error { // esta funcion escribe reportes de texto
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("no se pudo crear carpeta del reporte: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("no se pudo escribir reporte: %w", err)
	}

	fmt.Printf("reporte generado: %s\n", outputPath)
	return nil
}

func porcentaje(value float64, total float64) float64 { // esta funcion calcula porcentaje evitando division entre cero
	if total == 0 {
		return 0
	}

	return (value / total) * 100
}

type inodoReporte struct { // esta estructura agrupa indice e inodo para reportes
	Index int32
	Inode estructuras.Inode
}

type entradaReporte struct { // esta estructura representa una entrada de carpeta para reportes
	Nombre string
	Inode  int32
}

func inodosUsados(file *os.File, sb estructuras.SuperBlock) ([]inodoReporte, error) { // esta funcion obtiene todos los inodos marcados como usados
	bitmap := make([]byte, sb.InodesCount)
	if _, err := file.ReadAt(bitmap, int64(sb.BmInodeStart)); err != nil {
		return nil, fmt.Errorf("no se pudo leer bitmap de inodos: %w", err)
	}

	var usados []inodoReporte
	for index, value := range bitmap {
		if value != '1' {
			continue
		}

		inode, err := leerInodoPorIndice(file, sb, int32(index))
		if err != nil {
			return nil, err
		}

		usados = append(usados, inodoReporte{Index: int32(index), Inode: inode})
	}

	return usados, nil
}

func entradasCarpeta(file *os.File, sb estructuras.SuperBlock, inode estructuras.Inode) ([]entradaReporte, error) { // esta funcion lista entradas de una carpeta
	var entradas []entradaReporte

	for pointerIndex := 0; pointerIndex < 12; pointerIndex++ {
		blockIndex := inode.Block[pointerIndex]
		if blockIndex == -1 {
			continue
		}

		block, err := leerBloqueCarpeta(file, sb, blockIndex)
		if err != nil {
			return nil, err
		}

		for _, content := range block.Content {
			if content.Inode == -1 {
				continue
			}
			entradas = append(entradas, entradaReporte{Nombre: utils.BytesToString(content.Name[:]), Inode: content.Inode})
		}
	}

	return entradas, nil
}

func escribirNodoBloqueCarpeta(dot *strings.Builder, index int32, block estructuras.FolderBlock) { // esta funcion escribe un bloque carpeta en graphviz
	dot.WriteString(fmt.Sprintf("block%d [label=<\n<table border='1' cellborder='1' cellspacing='0'>\n", index))
	dot.WriteString(fmt.Sprintf("<tr><td colspan='2'><b>BLOQUE CARPETA %d</b></td></tr>\n", index))
	for _, content := range block.Content {
		dot.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>\n", html.EscapeString(utils.BytesToString(content.Name[:])), content.Inode))
	}
	dot.WriteString("</table>>];\n")
}

func escribirNodoBloqueArchivo(dot *strings.Builder, index int32, block estructuras.FileBlock) { // esta funcion escribe un bloque archivo en graphviz
	content := html.EscapeString(utils.BytesToString(block.Content[:]))
	dot.WriteString(fmt.Sprintf("block%d [label=<\n<table border='1' cellborder='1' cellspacing='0'>\n", index))
	dot.WriteString(fmt.Sprintf("<tr><td><b>BLOQUE ARCHIVO %d</b></td></tr>\n", index))
	dot.WriteString(fmt.Sprintf("<tr><td>%s</td></tr>\n", content))
	dot.WriteString("</table>>];\n")
}

func escribirTreeInodo(file *os.File, sb estructuras.SuperBlock, inodeIndex int32, dot *strings.Builder, visitados map[int32]bool) error { // esta funcion recorre inodos para reporte tree
	if visitados[inodeIndex] {
		return nil
	}
	visitados[inodeIndex] = true

	inode, err := leerInodoPorIndice(file, sb, inodeIndex)
	if err != nil {
		return err
	}

	tipo := "archivo"
	if inode.Type == estructuras.FolderBlockType {
		tipo = "carpeta"
	}

	dot.WriteString(fmt.Sprintf("inode%d [label=<\n<table border='1' cellborder='1' cellspacing='0'>\n", inodeIndex))
	dot.WriteString(fmt.Sprintf("<tr><td><b>INODO %d</b></td></tr><tr><td>%s</td></tr><tr><td>size=%d</td></tr>\n", inodeIndex, tipo, inode.Size))
	dot.WriteString("</table>>];\n")

	for _, blockIndex := range inode.Block {
		if blockIndex == -1 {
			continue
		}

		dot.WriteString(fmt.Sprintf("inode%d -> block%d;\n", inodeIndex, blockIndex))

		if inode.Type == estructuras.FolderBlockType {
			block, err := leerBloqueCarpeta(file, sb, blockIndex)
			if err != nil {
				return err
			}
			escribirNodoBloqueCarpeta(dot, blockIndex, block)

			for _, content := range block.Content {
				name := utils.BytesToString(content.Name[:])
				if content.Inode == -1 || name == "." || name == ".." {
					continue
				}

				dot.WriteString(fmt.Sprintf("block%d -> inode%d [label=\"%s\"];\n", blockIndex, content.Inode, html.EscapeString(name)))
				if err := escribirTreeInodo(file, sb, content.Inode, dot, visitados); err != nil {
					return err
				}
			}
		} else {
			var block estructuras.FileBlock
			if err := utils.ReadStructAt(file, int64(sb.BlockStart+blockIndex*sb.BlockSize), &block); err != nil {
				return err
			}
			escribirNodoBloqueArchivo(dot, blockIndex, block)
		}
	}

	return nil
}
