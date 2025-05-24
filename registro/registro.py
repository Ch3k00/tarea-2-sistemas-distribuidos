import pika
import json
from pymongo import MongoClient

# Conexión a MongoDB
mongo = MongoClient("mongodb://localhost:27017/")
db = mongo["emergencias"]
coleccion = db["emergencias"]

# Conexión a RabbitMQ
connection = pika.BlockingConnection(pika.ConnectionParameters("localhost"))
channel = connection.channel()

# Declarar colas
channel.queue_declare(queue="registro_q")
channel.queue_declare(queue="registro_fin_q")

# Callback para registrar emergencias nuevas
def manejar_emergencia_en_curso(ch, method, properties, body):
    emergencia = json.loads(body)
    emergencia["status"] = "En curso"

    # Insertar o actualizar por nombre (simplificación)
    coleccion.update_one(
        {"name": emergencia["name"]},
        {"$set": emergencia},
        upsert=True
    )
    print(f"📝 Registrada emergencia en curso: {emergencia['name']}")

# Callback para marcar emergencias como apagadas
def manejar_emergencia_extinguida(ch, method, properties, body):
    emergencia = json.loads(body)
    coleccion.update_one(
        {"name": emergencia["name"]},
        {"$set": {"status": "Extinguido"}}
    )
    print(f"✅ Emergencia extinguida: {emergencia['name']}")

# Consumir mensajes
channel.basic_consume(queue="registro_q", on_message_callback=manejar_emergencia_en_curso, auto_ack=True)
channel.basic_consume(queue="registro_fin_q", on_message_callback=manejar_emergencia_extinguida, auto_ack=True)

print("📚 Servicio de registro escuchando...")
channel.start_consuming()
