# -*- coding: utf-8 -*-

import os.path
import sys

loaded_modules = []


def load(path):
    """
    Loads the module from the specified path.

    :param path: The path to load the module of
    :return: The loaded module
    """
    basename = os.path.basename(path)
    module_name, extension = os.path.splitext(basename)
    plugin = load_module(module_name, path)
    loaded_modules.append(plugin)


if sys.version_info >= (3, 5):
    import importlib.util


    def load_module(module_name, path):
        """
        Loads the module with the specified name from the path.

        :param module_name: The name of the module to load
        :param path: The path to load the module of
        :return: The loaded module
        """
        spec = importlib.util.spec_from_file_location(module_name, path)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        return module
elif sys.version_info >= (3, 3):
    from importlib.machinery import SourceFileLoader


    def load_module(module_name, path):
        """
        Loads the module with the specified name from the path.

        :param module_name: The name of the module to load
        :param path: The path to load the module of
        :return: The loaded module
        """
        return SourceFileLoader(module_name, path).load_module()
