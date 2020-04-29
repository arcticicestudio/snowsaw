# -*- coding: utf-8 -*-

import os
import subprocess
import snowsaw
from socket import gethostname



class Shell(snowsaw.Plugin):
    """
    Core plugin to run arbitrary shell commands.
    """
    _directive = "shell"

    def can_handle(self, directive):
        return directive == self._directive

    def handle(self, directive, data):
        if directive != self._directive:
            raise ValueError("Core plugin \"Shell\" cannot handle the directive \"{}\"".format(directive))
        return self._process_commands(data)

    def _process_commands(self, data):
        """
        Processes specified commands.

        :param data: The commands to process
        :return: True if the commands have been processed successfully, False otherwise
        """
        success = True
        defaults = self._context.defaults().get("shell", {})
        with open(os.devnull, "w") as devnull:
            for item in data:
                stdin = stdout = stderr = devnull
                if isinstance(item, dict):
                    cmd = item["command"]
                    msg = item.get("description", None)
                    if item.get("stdin", defaults.get("stdin", False)) is True:
                        stdin = None
                    if item.get("stdout", defaults.get("stdout", False)) is True:
                        stdout = None
                    if item.get("stderr", defaults.get("stderr", False)) is True:
                        stderr = None
                    host = item.get("host", None)
                elif isinstance(item, list):
                    cmd = item[0]
                    msg = item[1] if len(item) > 1 else None
                else:
                    cmd = item
                    msg = None
                executable = os.environ.get("SHELL")
                shouldExecute = False
                if host is None:
                    shouldExecute = True
                elif host == gethostname():
                    shouldExecute = True
                elif host == "-":
                    shouldExecute = True
                if shouldExecute:
                    if msg is None:
                        self._log.lowinfo(cmd)
                    else:
                        self._log.lowinfo('{} [{}]'.format(msg, cmd))    
                    ret = subprocess.call(cmd, shell=True, stdin=stdin, stdout=stdout, stderr=stderr, cwd=self._context.snowblock_dir(),
                                      executable=executable)
                    if ret != 0:
                        success = False
                        self._log.warning("Command [{}] failed".format(cmd))
                else:
                    self._log.lowinfo("Skipping command [{}]".format(cmd))
        if success:
            self._log.info("=> All commands have been executed")
        else:
            self._log.error("Some commands were not successfully executed")
        return success
