from enum import Enum
from time import sleep, time
import uuid

# TODO: 
# - Usar evento o algun valor compartido entre hilos para llevar el estado de la tarea 

_tasks: dict[str, dict] = {}
_EXPIRATION_TIME_SECS = 10 * 60 # 10min
_alive_scheduler = False

class Status(str, Enum):
    in_progress = 'in_progress'
    completed = 'completed'
    error = 'error'
    expired = 'expired'

def init_task_scheduler():
    global _alive_scheduler
    _alive_scheduler = True
    while(_alive_scheduler):
        for id in _tasks:
            if time() - _tasks[id]["init_at"] >= _EXPIRATION_TIME_SECS:
                if _tasks[id]["status"] == Status.in_progress:
                    expire_task(id)
        sleep(5)

def end_task_scheduler():
    global _alive_scheduler
    _alive_scheduler = False

def expire_task(task_id):
    if task_id in _tasks: # Chequeamos porque otro hilo podría haberla borrado
        _tasks[task_id]["status"] = Status.expired
    return

def remove_task(task_id):
    if task_id not in _tasks:
        return

    del _tasks[task_id]

def new_task():
    task_id = str(uuid.uuid4())
    _tasks[task_id] = {
        "status": Status.in_progress,
        "task_id": task_id,
        "init_at": time()
    }

    print(_tasks)
    return task_id

def get_task(task_id: str) -> dict | None:
    return _tasks.get(task_id, None)

# TODO: maquina de estados
def update_status(task_id: str, status: Status):
    current_status = _tasks[task_id]["status"]

    # No se puede cambiar el 'status' a 'expired'
    if status == Status.expired: return

    # Solamete si el 'status' es 'in_progress' se puede cambiar
    if current_status == Status.in_progress:
        _tasks[task_id]["status"] = status

def update(task_id: str, **kwargs):
    _tasks[task_id] = {**_tasks[task_id], **kwargs}
