#!/usr/bin/env python3
import logging
import logging.handlers
import os

class LoggingModule:
    _instance = None

    def __new__(cls, *args, **kwargs):
        if not cls._instance:
            cls._instance = super(LoggingModule, cls).__new__(cls, *args, **kwargs)
            cls._instance._initialize()
        return cls._instance

    def _initialize(self):
        log_format = '%(asctime)s.%(msecs)03d %(levelname)s %(message)s'
        date_format = '%Y-%m-%d - %H:%M:%S'
        filename = 'portwhine.log'
        log_level = os.getenv('LOG_LEVEL', 'INFO').upper()

        # Convert log level from string to logging level
        log_level = getattr(logging, log_level, logging.INFO)

        root_logger = logging.getLogger()
        handler = logging.handlers.RotatingFileHandler(filename, 'a', 10**6, 10)
        formatter = logging.Formatter(fmt=log_format, datefmt=date_format)
        handler.setFormatter(formatter)
        root_logger.addHandler(handler)

        logger = logging.getLogger('portwhine')
        logger.setLevel(log_level)
        logger.info('Logging started (level=%s, filename=%s)',
                    logging.getLevelName(logger.getEffectiveLevel()), filename)

        print(f'See logfile for details: {os.path.join(os.getcwd(), filename)}')