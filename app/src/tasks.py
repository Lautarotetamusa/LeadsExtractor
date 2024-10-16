import uuid

# Private
_tasks: dict[str, dict] = {}

def new_task():
    # Generar un ID único para la tarea
    task_id = str(uuid.uuid4())
    print(task_id)
    _tasks[task_id] = {
        "status": "in_progress",
        "task_id": task_id
    }
    print(_tasks)
    return task_id

def get_task(task_id: str) -> dict | None:
    return _tasks.get(task_id, None)

# TODO: maquina de estados
def update_status(task_id: str, status):
    _tasks[task_id]["status"] = status

def update(task_id: str, **kwargs):
    _tasks[task_id] = {**_tasks[task_id], **kwargs}
