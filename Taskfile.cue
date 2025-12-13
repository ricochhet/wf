builtin: {
	cwd:  *"." | string
	time: *"2006-01-02_15-04-05.000" | string
}

_versions: {
	pwshVersion:                 "7.5.4"
	gitVersion:                  "2.52.0"
	nodeVersion:                 "25.2.1"
	bunVersion:                  "1.3.4"
	mongodVersion:               "8.2.2"
	mongoshVersion:              "2.5.10"
	mongodbDatabaseToolsVersion: "100.13.0"
	compassVersion:              "1.48.2"
}

_binaries: {
	pwshBinary:                 "PowerShell-" + _versions.pwshVersion + "-win-x64"
	gitBinary:                  "PortableGit-" + _versions.gitVersion + "-64-bit.7z"
	nodeBinary:                 "node-v" + _versions.nodeVersion + "-win-x64"
	bunBinary:                  "bun-windows-x64"
	mongoshBinary:              "mongosh-" + _versions.mongoshVersion + "-win32-x64"
	mongodbDatabaseToolsBinary: "mongodb-database-tools-windows-x86_64-" + _versions.mongodbDatabaseToolsVersion
	compassBinary:              "mongodb-compass-" + _versions.compassVersion + "-win32-x64"
}

_commands: {
	pwsh:             builtin.cwd + "\\OpenWF\\Boot\\" + _binaries.pwshBinary + "\\pwsh.exe"
	git:              builtin.cwd + "\\OpenWF\\Boot\\" + _binaries.gitBinary + "\\bin\\git.exe"
	mkdir:            builtin.cwd + "\\OpenWF\\Boot\\" + _binaries.gitBinary + "\\usr\\bin\\mkdir.exe"
	npm:              builtin.cwd + "\\OpenWF\\Boot\\" + _binaries.nodeBinary + "\\" + _binaries.nodeBinary + "\\npm.cmd"
	mongod:           builtin.cwd + "\\OpenWF\\Boot\\mongodb-windows-x86_64-" + _versions.mongodVersion + "\\mongodb-win32-x86_64-windows-" + _versions.mongodVersion + "\\bin\\mongod.exe"
	mongosh:          builtin.cwd + "\\OpenWF\\Boot\\" + _binaries.mongoshBinary + "\\" + _binaries.mongoshBinary + "\\bin\\mongosh.exe"
	mongodump:        builtin.cwd + "\\OpenWF\\Boot\\" + _binaries.mongodbDatabaseToolsBinary + "\\" + _binaries.mongodbDatabaseToolsBinary + "\\bin\\mongodump.exe"
	mongorestore:     builtin.cwd + "\\OpenWF\\Boot\\" + _binaries.mongodbDatabaseToolsBinary + "\\" + _binaries.mongodbDatabaseToolsBinary + "\\bin\\mongorestore.exe"
	compass:          builtin.cwd + "\\OpenWF\\Boot\\" + _binaries.compassBinary + "\\MongoDBCompass.exe"
	irc:              builtin.cwd + "\\OpenWF\\Boot\\IRC\\IRC\\inspircd.exe"
	warframe:         builtin.cwd + "\\Warframe\\Warframe.x64.exe"
	warframeLauncher: builtin.cwd + "\\Warframe\\Tools\\Launcher.exe"
}

env: PATH: [
	builtin.cwd + "\\OpenWF\\Boot\\" + _binaries.nodeBinary + "\\" + _binaries.nodeBinary + "\\",
	builtin.cwd + "\\OpenWF\\Boot\\" + _binaries.bunBinary + "-" + _versions.bunVersion + "\\" + _binaries.bunBinary + "\\",
]

runas: [{
	name: "SNS.Server.exe"
	tasks: [
		"database",
		"server",
		"irc",
	]
	port:          8555
	interval:      2
	reverseOnStop: true
	start:         true
}, {
	name: "SNS.IRC.exe"
	tasks: ["irc"]
	port:  8556
	start: true
}, {
	name: "Warframe.DX11.exe"
	tasks: ["warframe:dx11"]
	port:  8557
	start: true
}, {
	name: "gpm.exe"
	tasks: ["console"]
	port: 8558
}]

tasks: [{
	name: "setup"
	cmd: ["gpm:pull"]
	steps: [
		"gpm:prune",
		"clone-server",
		"update-server",
		"bootstrap",
		"dbdata",
	]
	dir: ""
	platforms: []
}, {
	name: "prune"
	cmd: ["gpm:prune"]
	steps: []
	dir: ""
	platforms: []
}, {
	name: "dbdata"
	cmd: [
		_commands.mkdir,
		"Data",
	]
	steps: []
	dir: "OpenWF"
	platforms: []
}, {
	name: "bootstrap"
	cmd: [
		_commands.pwsh,
		"-ExecutionPolicy",
		"Bypass",
		"-File",
		"Download Latest DLL.ps1",
	]
	steps: []
	dir: "Warframe/OpenWF"
	platforms: []
}, {
	name: "clone-server"
	cmd: [
		_commands.git,
		"clone",
		"--recursive",
		"https://onlyg.it/OpenWF/SpaceNinjaServer",
	]
	steps: [
		"clone-server-data",
		"build-server",
	]
	dir: "OpenWF"
	platforms: []
}, {
	name: "clone-server-data"
	cmd: [
		_commands.git,
		"clone",
		"--recursive",
		"https://openwf.io/0.git",
	]
	steps: []
	dir: "OpenWF/SpaceNinjaServer/static/data"
	platforms: []
}, {
	name: "update-server"
	cmd: [
		_commands.git,
		"fetch",
		"--prune",
	]
	steps: [
		"update-server:stash",
		"update-server:checkout",
		"update-server:pull",
		"build-server",
	]
	dir: "OpenWF/SpaceNinjaServer"
	platforms: []
}, {
	name: "update-server:stash"
	cmd: [
		_commands.git,
		"stash",
	]
	steps: []
	dir: "OpenWF/SpaceNinjaServer"
	platforms: []
}, {
	name: "update-server:checkout"
	cmd: [
		_commands.git,
		"checkout",
		"-f",
		"origin/main",
	]
	steps: []
	dir: "OpenWF/SpaceNinjaServer"
	platforms: []
}, {
	name: "update-server:pull"
	cmd: [
		_commands.git,
		"pull",
	]
	steps: []
	dir: "OpenWF/SpaceNinjaServer/static/data/0"
	platforms: []
}, {
	name: "build-server"
	cmd: [
		_commands.npm,
		"i",
		"--omit=dev",
	]
	steps: ["build-server:build"]
	dir: "OpenWF/SpaceNinjaServer"
	platforms: []
}, {
	name: "build-server:build"
	cmd: [
		_commands.npm,
		"run",
		"build",
	]
	steps: []
	dir: "OpenWF/SpaceNinjaServer"
	platforms: []
}, {
	name: "server"
	cmd: [
		_commands.npm,
		"run",
		"start",
	]
	steps: []
	dir: "OpenWF/SpaceNinjaServer"
	platforms: []
}, {
	name: "server:bun"
	cmd: [
		_commands.npm,
		"run",
		"bun-run",
	]
	steps: []
	dir: "OpenWF/SpaceNinjaServer"
	platforms: []
}, {
	name: "database"
	cmd: [
		_commands.mongod,
		"--dbpath",
		"Data",
	]
	steps: []
	dir: "OpenWF"
	platforms: []
}, {
	name: "database:backup"
	cmd: [
		_commands.mkdir,
		"backups",
	]
	steps: ["database:backup2"]
	dir: "OpenWF"
	platforms: []
}, {
	name: "database:backup2"
	cmd: [
		_commands.mongodump,
		"--uri=mongodb://127.0.0.1:27017/openWF",
		"--gzip",
		"--archive=DataBackup-" + builtin.time + ".gz",
	]
	steps: []
	dir: "OpenWF/backups"
	platforms: []
}, {
	name: "database:restore"
	cmd: [
		_commands.mongosh,
		"--quiet",
		"--eval",
		"db.dropDatabase();",
		"mongodb://127.0.0.1:27017/openWF",
	]
	steps: ["database:restore2"]
	dir: "OpenWF"
	platforms: []
}, {
	name: "database:restore2"
	cmd: [
		_commands.mongorestore,
		"--uri=mongodb://127.0.0.1:27017/openWF",
		"--gzip",
		"--archive=DataBackup.gz",
	]
	steps: []
	dir: "OpenWF"
	platforms: []
}, {
	name: "compass"
	cmd: [_commands.compass]
	steps: []
	dir: "OpenWF"
	platforms: []
}, {
	name: "irc"
	cmd: [
		_commands.irc,
		"--nofork",
	]
	steps: []
	dir: "OpenWF/Boot/IRC/IRC"
	platforms: []
}, {
	name: "warframe:launcher"
	cmd: [
		_commands.warframeLauncher,
		"-cluster:public",
		"-registry:Steam",
	]
	steps: []
	dir: "Warframe/Tools"
	platforms: []
}, {
	name: "warframe:dx11"
	cmd: [
		_commands.warframe,
		"-fullscreen:0",
		"-shaderCache:1",
		"-graphicsDriver:dx11",
		"-gpuPreference:2",
		"-language:en",
		"-languageVO:en",
	]
	steps: []
	dir:  "Warframe"
	fork: true
	platforms: []
}, {
	name: "warframe:dx12"
	cmd: [
		_commands.warframe,
		"-fullscreen:0",
		"-shaderCache:1",
		"-graphicsDriver:dx12",
		"-gpuPreference:2",
		"-language:en",
		"-languageVO:en",
	]
	steps: []
	dir:  "Warframe"
	fork: true
	platforms: []
}, {
	name: "warframe:dx11:defrag"
	cmd: [
		_commands.warframe,
		"-silent",
		"-log:/Defrag.log",
		"-graphicsDriver:dx11",
		"-cluster:public",
		"-language:en",
		"-applet:/EE/Types/Framework/CacheDefraggerIOCP",
		"/Tools/CachePlan.txt",
	]
	steps: ["warframe:dx11:contentupdate"]
	dir: "Warframe"
	platforms: []
}, {
	name: "warframe:dx12:defrag"
	cmd: [
		_commands.warframe,
		"-silent",
		"-log:/Defrag.log",
		"-graphicsDriver:dx12",
		"-cluster:public",
		"-language:en",
		"-applet:/EE/Types/Framework/CacheDefraggerIOCP",
		"/Tools/CachePlan.txt",
	]
	steps: ["warframe:dx12:contentupdate"]
	dir: "Warframe"
	platforms: []
}, {
	name: "warframe:dx11:repair"
	cmd: [
		_commands.warframe,
		"-silent",
		"-log:/Repair.log",
		"-graphicsDriver:dx11",
		"-cluster:public",
		"-language:en",
		"-applet:/EE/Types/Framework/CacheRepair",
	]
	steps: ["warframe:dx11:contentupdate"]
	dir: "Warframe"
	platforms: []
}, {
	name: "warframe:dx12:repair"
	cmd: [
		_commands.warframe,
		"-silent",
		"-log:/Repair.log",
		"-graphicsDriver:dx12",
		"-cluster:public",
		"-language:en",
		"-applet:/EE/Types/Framework/CacheRepair",
	]
	steps: ["warframe:dx12:contentupdate"]
	dir: "Warframe"
	platforms: []
}, {
	name: "warframe:dx11:contentupdate"
	cmd: [
		_commands.warframe,
		"-silent",
		"-log:/Preprocess.log",
		"-graphicsDriver:dx11",
		"-cluster:public",
		"-language:en",
		"-applet:/EE/Types/Framework/ContentUpdate",
	]
	steps: []
	dir: "Warframe"
	platforms: []
}, {
	name: "warframe:dx12:contentupdate"
	cmd: [
		_commands.warframe,
		"-silent",
		"-log:/Preprocess.log",
		"-graphicsDriver:dx12",
		"-cluster:public",
		"-language:en",
		"-applet:/EE/Types/Framework/ContentUpdate",
	]
	steps: []
	dir: "Warframe"
	platforms: []
}]
artifacts: {
	pull: [{
		url:     "https://github.com/PowerShell/PowerShell/releases/download/v" + _versions.pwshVersion + "/" + _binaries.pwshBinary + ".zip"
		dir:     "OpenWF/Downloads/"
		extract: "OpenWF/Boot/"
		platforms: []
	}, {
		url:     "https://github.com/git-for-windows/git/releases/download/v" + _versions.gitVersion + ".windows.1/" + _binaries.gitBinary + ".exe"
		dir:     "OpenWF/Downloads/"
		extract: "OpenWF/Boot/"
		platforms: []
		force: true
	}, {
		url:     "https://nodejs.org/dist/v" + _versions.nodeVersion + "/" + _binaries.nodeBinary + ".zip"
		dir:     "OpenWF/Downloads/"
		extract: "OpenWF/Boot/"
		platforms: []
	}, {
		url:      "https://github.com/oven-sh/bun/releases/download/bun-v" + _versions.bunVersion + "/" + _binaries.bunBinary + ".zip"
		dir:      "OpenWF/Downloads/"
		filename: _binaries.bunBinary + "-" + _versions.bunVersion + ".zip"
		extract:  "OpenWF/Boot/"
		platforms: []
	}, {
		url:     "https://fastdl.mongodb.org/windows/mongodb-windows-x86_64-" + _versions.mongodVersion + ".zip"
		dir:     "OpenWF/Downloads/"
		extract: "OpenWF/Boot/"
		platforms: []
	}, {
		url:     "https://fastdl.mongodb.org/tools/db/" + _binaries.mongodbDatabaseToolsBinary + ".zip"
		dir:     "OpenWF/Downloads/"
		extract: "OpenWF/Boot/"
		platforms: []
	}, {
		url:     "https://downloads.mongodb.com/compass/" + _binaries.mongoshBinary + ".zip"
		dir:     "OpenWF/Downloads/"
		extract: "OpenWF/Boot/"
		platforms: []
	}, {
		url:     "https://github.com/mongodb-js/compass/releases/download/v" + _versions.compassVersion + "/" + _binaries.compassBinary + ".zip"
		dir:     "OpenWF/Downloads/"
		extract: "OpenWF/Boot/"
		platforms: []
	}, {
		url:     "https://openwf.io/supplementals/IRC.zip"
		dir:     "OpenWF/Downloads/"
		extract: "OpenWF/Boot/"
		platforms: []
	}, {
		url:     "https://openwf.io/supplementals/Download%20Latest%20DLL.ps1"
		dir:     "Warframe/OpenWF/"
		extract: "OpenWF/Boot/"
		platforms: []
	}]
	prune: []
}
