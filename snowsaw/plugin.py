# -*- coding: utf-8 -*-

from .logging import Logger
from .context import Context


class Plugin(object):
    """
    A abstract base class for plugins that process directives.
    """
    def __init__(self, context):
        self._context = context
        self._log = Logger()

    def can_handle(self, directive):
        """
        Checks if the plugin can handle the specified directive.

        :param directive: The directive to check
        :return: True if the specified directive can be handled, False otherwise
        """
        raise NotImplementedError

    def handle(self, directive, data):
        """
        Handles the data of the specified directive.

        :param directive: The directive to handle the data of
        :param data: The data to handle
        :return: True if the directive has been handled successfully
        """
        raise NotImplementedError
