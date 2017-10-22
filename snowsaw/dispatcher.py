# -*- coding: utf-8 -*-

import os
from .plugin import Plugin
from .logging import Logger
from .context import Context


class Dispatcher(object):
    """
    Dispatches tasks to loaded plugins.
    """
    def __init__(self, snowblock_dir):
        self._log = Logger()
        self._setup_context(snowblock_dir)
        self._load_plugins()

    def _setup_context(self, snowblock_dir):
        """
        Sets up the plugin context for the specified snowblock.

        :param snowblock_dir:  The directory of the snowblock
        :return: The plugin context
        """
        path = os.path.abspath(os.path.realpath(os.path.expanduser(snowblock_dir)))
        if not os.path.exists(path):
            raise DispatchError("Nonexistent snowblock directory")
        self._context = Context(path)

    def dispatch(self, tasks):
        """
        Dispatches the specified tasks to the loaded plugins.

        :param tasks: The tasks to dispatch
        :return: True if all tasks have been handled by the loaded plugins successfully, False otherwise
        """
        success = True
        for task in tasks:
            for action in task:
                handled = False
                if action == "defaults":
                    self._context.set_defaults(task[action])
                    handled = True
                for plugin in self._plugins:
                    if plugin.can_handle(action):
                        try:
                            success &= plugin.handle(action, task[action])
                            handled = True
                        except Exception:
                            self._log.error("An error was encountered while executing action \"{}\"".format(action))
                if not handled:
                    success = False
                    self._log.error("Action \"{}\" not handled".format(action))
        return success

    def _load_plugins(self):
        """
        Loads all found plugins.

        :return: None
        """
        self._plugins = [plugin(self._context) for plugin in Plugin.__subclasses__()]


class DispatchError(Exception):
    pass
