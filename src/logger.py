from enum import Enum
from time import gmtime, strftime

LOG_LEVEL = 1 #OLNY FOR STD; FILE LOGS PRINTS ALL; 2->error, 3->success, 1->debug

class LogType(Enum):
    debug = 1
    error = 2
    warning = 3
    success = 4

class Logger():
    def __init__(self, fuente):
        self.fuente = fuente

    def log_print(self, msg: str, log_type: LogType):
        if log_type == LogType.debug:
            color = "\033[0m"
        elif log_type == LogType.error:
            color = "\033[91m"
        elif log_type == LogType.success:
            color = "\033[92m"
        
        print(f"\033[95m[{strftime('%H:%M:%S', gmtime())}][{self.fuente}]{color} {msg}")

    def debug(self, msg):
        self.log_print(msg, LogType.debug)

    def error(self, msg):
        self.log_print(msg, LogType.error)

    def warning(self, msg):
        self.log_print(msg, LogType.warning)

    def success(self, msg):
        self.log_print(msg, LogType.success)