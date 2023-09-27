package gom

type _FileMigration struct {
	name        string
	fileName    string
	fileContent string
}

type _DbMigration struct {
	Name string `db:"name"`
}

const (
	DefaultMigrationsDir = "migrations"
)

const (
	upMigrationType   = iota
	downMigrationType = iota
	upComment         = "-- migration:up"
	downComment       = "-- migration:down"
)

var baseFS simpleFS = gomFS{}
var migrationsDir = DefaultMigrationsDir

func SetBaseFS(fsys simpleFS) {
	if fsys != nil {
		baseFS = fsys
	}
}

func SetMigrationsDir(dir string) {
	if dir == "" {
		migrationsDir = DefaultMigrationsDir
	} else {
		migrationsDir = dir
	}
}
