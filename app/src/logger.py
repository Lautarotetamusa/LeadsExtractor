from enum import Enum
from time import gmtime, strftime

class LogType(Enum):
    success = 1
    error = 2
    warning = 3
    debug = 4

LOG_LEVEL = LogType.debug 
CHAT_LOG_LEVEL = LogType.error
GOOGLE_CHAT_ENABLE = False
COLOR_MAP = {
    LogType.error: {
        "color": "\033[91m",
    },
    LogType.success: {
        "color": "\033[92m",
    },
    LogType.warning: {
        "color": "\033[93m",
    },
    LogType.debug: {
        "color": "\033[39m",
    }
}

class Logger():
    def __init__(self, fuente):
        self.fuente = fuente

    def log_print(self, msg: str, log_type: LogType):
        color = COLOR_MAP[log_type]["color"]
        time = strftime('%H:%M:%S', gmtime())

        if log_type.value <= LOG_LEVEL.value:
            print(f"\033[95m[{time}][{self.fuente}]{color} {msg}")

    def debug(self, msg):
        self.log_print(msg, LogType.debug)

    def error(self, msg):
        self.log_print(msg, LogType.error)

    def warning(self, msg):
        self.log_print(msg, LogType.warning)

    def success(self, msg):
        self.log_print(msg, LogType.success)
