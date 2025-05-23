## ⚙️ Requisitos y Configuración del Entorno

Este proyecto requiere configurar un entorno con Go, Python, gRPC, RabbitMQ y MongoDB para poder desarrollar y ejecutar los servicios distribuidos correctamente.

---

### 🔧 Dependencias necesarias

- Go 1.20 o superior
- Python 3.10 o superior
- RabbitMQ
- MongoDB (v5 o superior)
- Protocol Buffers Compiler (`protoc`)
- Plugins gRPC para Go y Python

---

### 🐹 Instalar Go

1. Descargar desde: https://go.dev/dl
2. Instalar y asegurarse de que `go` esté en el `PATH`.
3. Verificar con:

```bash
go version
```

---

### 🐍 Instalar Python + gRPC para Python

```bash
pip install grpcio grpcio-tools
```

---

### 🐰 Instalar RabbitMQ

1. Instalar Erlang desde: https://www.erlang.org/downloads  
2. Instalar RabbitMQ desde: https://www.rabbitmq.com/install-windows.html  
3. Habilitar el panel web:

```bash
rabbitmq-plugins enable rabbitmq_management
```

Acceder al panel desde [http://localhost:15672](http://localhost:15672)  
Usuario: `guest` | Clave: `guest`

---

### 🍃 Instalar MongoDB

1. Descargar e instalar desde: https://www.mongodb.com/try/download/community  
2. Asegurarse de que el servicio `mongod` esté corriendo.  
3. Verificar con:

```bash
mongosh
```

---

### 🧩 Instalar Protocol Buffers Compiler (`protoc`)

1. Descargar desde: https://github.com/protocolbuffers/protobuf/releases  
2. Extraer y agregar la carpeta `bin/` al `PATH`.  
3. Verificar con:

```bash
protoc --version
```

---

### 🚀 Configurar el módulo Go del proyecto

Desde la raíz del proyecto, inicializar el módulo:

```bash
go mod init github.com/tuusuario/tarea-2-sd
go get google.golang.org/grpc
```

Reemplaza `tuusuario` con tu nombre de usuario o el nombre de tu organización.

---

### 🛠️ Compilar el archivo `.proto`

**Para Go:**

```bash
protoc --go_out=. --go-grpc_out=. proto/emergencias.proto
```

**Para Python (dentro de la carpeta raíz):**

```bash
python -m grpc_tools.protoc -Iproto --python_out=registro --grpc_python_out=registro proto/emergencias.proto
```

Una vez configurado el entorno, ya puedes compilar y ejecutar los servicios siguiendo las instrucciones del proyecto.
