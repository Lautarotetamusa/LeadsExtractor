package onedrive

type Path struct {
}

type OneDriver interface {
    Pwd() Path
    Ls() []string // list of folder names
    Cd(folder string) (Path, error)
}
