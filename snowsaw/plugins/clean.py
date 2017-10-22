# -*- coding: utf-8 -*-

import snowsaw
import os


class Clean(snowsaw.Plugin):
    """
    Core plugin to clean broken symbolic links.
    """
    _directive = "clean"

    def can_handle(self, directive):
        return directive == self._directive

    def handle(self, directive, data):
        if directive != self._directive:
            raise ValueError("Core plugin \"Clean\" cannot handle the directive \"{}\"".format(directive))
        return self._process_clean(data)

    def _process_clean(self, targets):
        """
        Processes specified targets.

        :param targets: The targets to process
        :return: True if the targets have been processed successfully
        """
        success = True
        for target in targets:
            success &= self._clean(target)
        if success:
            self._log.info("=> All targets have been cleaned")
        else:
            self._log.error("Some targets were not successfully cleaned")
        return success

    def _clean(self, target):
        """
        Cleans all broken symbolic links in the specified target that point to a subdirectory of the snowblocks directory.

        :param target: The target to clean
        :return: True if cleaned successfully
        """
        if not os.path.isdir(os.path.expanduser(target)):
            self._log.debug("Ignoring nonexistent directory {}".format(target))
            return True
        for item in os.listdir(os.path.expanduser(target)):
            path = os.path.join(os.path.expanduser(target), item)
            if not os.path.exists(path) and os.path.islink(path):
                if self._in_directory(path, self._context.snowblock_dir()):
                    self._log.lowinfo("Removing invalid link {} -> {}".format(path, os.path.join(os.path.dirname(path), os.readlink(path))))
                    os.remove(path)
        return True

    def _in_directory(self, path, directory):
        """
        Checks if the specified path is in the directory.

        :param path: The path to check
        :param directory: The directory to get checked
        :return: True if the path is in the directory
        """
        directory = os.path.join(os.path.realpath(directory), "")
        path = os.path.realpath(path)
        return os.path.commonprefix([path, directory]) == directory
