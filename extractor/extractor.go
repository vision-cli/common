package extractor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vision-cli/common/transpiler/model"
)

// type Module struct {
// 	ApiVersion string    `yaml:"apiVersion"`
// 	Name       string    `yaml:"name"`
// 	Services   []Service `yaml:"services"`
// }

// service := model.Service{
// 	Name: "projects",
// 	Enums: []model.Enum{
// 		{Name: "project-type", Values: []string{"not-assigned", "internal", "billable"}},
// 	},
// 	Entities: []model.Entity{
// 		{
// 			Name:        "project",
// 			Persistence: "db",
// 			Fields: []model.Field{
// 				{Name: "name", Type: "string", Tag: "db:", IsArray: false, IsNullable: true, IsSearchable: false},
// 			},
// 		},
// 	},
// }

func GetProjectStructure(projectDirectory string) string {
	var result string
	targetDir := filepath.Join(projectDirectory, "services")

	//Getting Module Names
	modules := getModules(targetDir)

	//Getting services names
	for i := range modules {
		moduleDirectory := filepath.Join(targetDir, modules[i].Name)
		modules[i].Services = getServices(moduleDirectory)
	}

	//This is just used for testing to print the structure we're actually creating, for verification purposes
	for i, module := range modules {
		result += fmt.Sprintf("\nModule Name is: "+module.Name+" - Module index is: %d\n", i)
		for j, service := range module.Services {
			result += fmt.Sprintf("\tService Name is: "+service.Name+" - Service index is: %d\n", j)
			for k, enum := range service.Enums {
				result += fmt.Sprintf("\t\tEnum Name is: "+enum.Name+" - Enum index is: %d\n", k)
				for l, value := range enum.Values {
					result += fmt.Sprintf("\t\t\tEnum Value is: "+value+" - Enum Value index is: %d\n", l)
				}
			}
		}
	}

	return "[]model.Module{}, []model.Service{}sameas: " + result
}

func getModules(targetDir string) []model.Module {
	moduleDirs, _ := os.ReadDir(targetDir)
	modules := []model.Module{}

	for _, path := range moduleDirs {
		if path.IsDir() && path.Name() != "default" {
			modules = append(modules, model.Module{Name: path.Name()})
		}
	}

	return modules
}

// service := model.Service{
// 	Name: "projects",
// 	Enums: []model.Enum{
// 		{Name: "project-type", Values: []string{"not-assigned", "internal", "billable"}},
// 	},
// 	Entities: []model.Entity{
// 		{
// 			Name:        "project",
// 			Persistence: "db",
// 			Fields: []model.Field{
// 				{Name: "name", Type: "string", Tag: "db:", IsArray: false, IsNullable: true, IsSearchable: false},
// 			},
// 		},
// 	},
// }

func getServices(moduleDirectory string) []model.Service {
	serviceDirs, _ := os.ReadDir(moduleDirectory)
	services := []model.Service{}

	for i, path := range serviceDirs {
		if path.IsDir() {
			services = append(services, model.Service{Name: path.Name()})
			// modelsFolder := filepath.Join(moduleDirectory, path.Name(), "models")
			services[i].Enums = getEnums("modelsFolder")
			// services[i].Enums = []model.Enum{{Name: "Test", Values: []string{"testing", "Enums"}}}
			// services[i].Entities = getEntities(modelsFolder)
		}
	}

	return services
}

func getEntities(modelsFolder string) []model.Entity {
	panic("unimplemented")
}

func getEnums(modelsFolder string) []model.Enum {
	// modelsGo := filepath.Join(modelsFolder, "models.go")
	enums := []model.Enum{}

	enums = append(enums, model.Enum{Name: "testing", Values: []string{"Temps"}})

	// Create a new file set
	// fset := token.NewFileSet()

	// Parse the file and retrieve the AST
	// file, err := parser.ParseFile(fset, modelsGo, nil, parser.ParseComments)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Process the AST as needed
	// Example: Print the names of all struct types
	// ast.Inspect(file, func(node ast.Node) bool {
	// 	if typeSpec, ok := node.(*ast.TypeSpec); ok {
	// 		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
	// 			enums[0].Values[0] = fmt.Sprint("Struct Name:", typeSpec.Name)
	// 			for _, field := range structType.Fields.List {
	// 				enums[0].Values[1] = fmt.Sprint("Field Name:", field.Names)
	// 			}
	// 		}
	// 	}
	// 	return true
	// })

	return enums
}
