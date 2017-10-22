# -*- coding: utf-8 -*-

import copy


class Context(object):
    """
    Contextual data and information for plugins
    """
    def __init__(self, snowblock_dir):
        self._snowblock_dir = snowblock_dir
        self._defaults = {}
        pass

    def set_snowblock_dir(self, snowblock_dir):
        self._snowblock_dir = snowblock_dir

    def snowblock_dir(self):
        return self._snowblock_dir

    def set_defaults(self, defaults):
        self._defaults = defaults

    def defaults(self):
        return copy.deepcopy(self._defaults)
