# BFI-44 Plugin

Big Five Inventory-44: Test de personalidad de los Cinco Grandes factores.

## Estructura

- `main.go` - Punto de entrada
- `plugin.go` - Struct e interface implementation
- `metadata.go` - NewBFI44Plugin() + buildBFI44Questions()
- `types.go` - Tipos de datos (Test, Question, Response, Answer, Results)
- `handlers.go` - CRUD de surveys/tests
- `responses.go` - Manejo de respuestas
- `calculator.go` - getBFI44Questions() + lógica de cálculo
- `bfi44_test.go` - Tests unitarios
- `example_bfi44_main.go` - Ejemplo de uso

## Dimensiones OCEAN

- **E** (Extraversion): 8 items
- **A** (Agreeableness): 9 items  
- **C** (Conscientiousness): 9 items
- **N** (Neuroticism): 8 items
- **O** (Openness): 10 items

## Compile

```
go build -o ../../plugins_dist/bfi44 .
```
