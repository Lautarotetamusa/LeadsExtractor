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

    def log_print(self, log_type: LogType, *args):
        color = COLOR_MAP[log_type]["color"]
        time = strftime('%H:%M:%S', gmtime())

        if log_type.value <= LOG_LEVEL.value:
            print(f"\033[95m[{time}][{self.fuente}]{color}", *args)

    def debug(self, *args):
        self.log_print(LogType.debug, *args)

    def error(self, *args):
        self.log_print(LogType.error, *args)

    def warning(self, *args):
        self.log_print(LogType.warning, *args)

    def success(self, *args):
        self.log_print(LogType.success, *args)
