package comandos

import (
	"fmt"           // uso fmt para armar textos de reportes
	"os"            // uso os para crear carpetas y archivos
	"os/exec"       // uso exec para llamar graphviz cuando sea posible
	"path/filepath" // uso filepath para crear carpetas y detectar extensiones
	"strings"       // uso strings para construir dot y normalizar nombres

	"MIA_P1_202243063/estructuras" // uso estructuras para leer datos binarios
	"MIA_P1_202243063/utils"       // uso utils para leer structs y limpiar bytes
)

func EjecutarREP(params map[string]string) error { // esta funcion ejecuta el comando rep
	mounted, exists := FindMountedPartitionByID(params["id"]) // busco la particion montada del reporte
	if !exists {
		return fmt.Errorf("no existe una particion montada con id %s", params["id"])
	}

	name := strings.ToLower(params["name"])
	switch name {
	case "mbr":
		return reporteMBR(*mounted, params["path"])
	case "disk":
		return reporteDISK(*mounted, params["path"])
	case "sb":
		return reporteSB(*mounted, params["path"])
	case "bm_inode":
		return reporteBitmap(*mounted, params["path"], true)
	case "bm_block":
		return reporteBitmap(*mounted, params["path"], false)
	case "file":
		return reporteFILE(*mounted, params["path"], params["path_file_ls"])
	default:
		return fmt.Errorf("reporte %s pendiente de implementar", name)
	}
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
	}

	dot.WriteString("</table>>];\n")
	dot.WriteString("}\n")

	return escribirGraphviz(outputPath, dot.String())
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

		dot.WriteString(fmt.Sprintf("<td>%s<br/>%c<br/>%.2f%%</td>\n", utils.BytesToString(partition.Name[:]), partition.Type, porcentaje(float64(partition.Size), float64(mbr.Size))))
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

	contenido, err := leerArchivoPorRuta(file, sb, filePath)
	if err != nil {
		return err
	}

	return escribirArchivoReporte(outputPath, fmt.Sprintf("archivo: %s\n%s\n", filePath, contenido))
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
		return os.WriteFile(outputPath, []byte(dot), 0644)
	}

	dotPath := outputPath + ".dot"
	if err := os.WriteFile(dotPath, []byte(dot), 0644); err != nil {
		return fmt.Errorf("no se pudo escribir dot: %w", err)
	}

	cmd := exec.Command("dot", "-T"+extension, dotPath, "-o", outputPath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("no se pudo renderizar con graphviz, se dejo dot en %s\n", dotPath)
		return nil
	}

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
