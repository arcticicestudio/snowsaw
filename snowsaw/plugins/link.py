# -*- coding: utf-8 -*-

import os
import shutil
import snowsaw
from socket import gethostname


class Link(snowsaw.Plugin):
    """
    Core plugin to symbolically link dotfiles.
    """
    _directive = "link"

    def can_handle(self, directive):
        return directive == self._directive

    def handle(self, directive, data):
        if directive != self._directive:
            raise ValueError("Core plugin \"Link\" cannot handle the directive \"{}\"".format(directive))
        return self._process_links(data)

    def _process_links(self, links):
        """
        Processes specified links.

        :param links: The links to process
        :return: True if the links have been processed successfully, False otherwise
        """
        success = True
        defaults = self._context.defaults().get("link", {})
        hostname = gethostname()
        for destination, source in links.items():
            destination = os.path.expandvars(destination)
            relative = defaults.get("relative", False)
            force = defaults.get("force", False)
            relink = defaults.get("relink", False)
            create = defaults.get("create", False)
            hosts = defaults.get("hosts", {})
            if isinstance(source, dict):
                relative = source.get("relative", relative)
                force = source.get("force", force)
                relink = source.get("relink", relink)
                create = source.get("create", create)
                path = self._default_source(destination, source.get("path"))
                hosts = source.get("hosts", hosts)
            else:
                path = self._default_source(destination, source)
            path = os.path.expandvars(os.path.expanduser(path))

            if hosts:
                if hostname in hosts:
                    path = os.path.expandvars(os.path.expanduser(hosts.get(hostname)))
                elif "-" in hosts:
                    path = os.path.expandvars(os.path.expanduser(hosts.get("-")))
                    self._log.lowinfo("Applying default link {} -> {}".format(destination, os.path.join(self._context.snowblock_dir(),path)))
                else:
                    for host in hosts.items():
                        self._log.lowinfo("Skipped host specific link {} -> {}".format(destination, os.path.join(self._context.snowblock_dir(), host[1])))
                    continue

            if not self._exists(os.path.join(self._context.snowblock_dir(), path)):
                success = False
                self._log.warning("Nonexistent target {} -> {}".format(destination, path))
                continue
            if create:
                success &= self._create(destination)
            if force or relink:
                success &= self._delete(path, destination, relative, force)
            success &= self._link(path, destination, relative)
        if success:
            self._log.info("=> All links have been set up")
        else:
            self._log.error("Some links were not successfully set up")
        return success

    def _default_source(self, destination, source):
        """
        Sets the default source if the value is empty, the configured source otherwise.

        If the source for a link is null, it uses the basename of the destination.
        Leading dot characters in the basename will be stripped.

        Allows simplified configuration files by avoiding unnecessary duplicate values.

        :param destination: The link destination
        :param source: The link source
        :return: The source string
        """
        if source is None:
            basename = os.path.basename(destination)
            if basename.startswith('.'):
                return basename[1:]
            else:
                return basename
        else:
            return source

    def _is_link(self, path):
        """
        Checks if the specified path is a symbolic link.

        :param path: The path to check
        :return: True if the path is a symbolic link
        """
        return os.path.islink(os.path.expanduser(path))

    def _link_destination(self, path):
        """
        Gets the destination of the specified symbolic link.

        :param path: The symbolic link to get the destination of
        :return: The symbolic link destination
        """
        path = os.path.expanduser(path)
        return os.readlink(path)

    def _exists(self, path):
        """
        Checks if the specified path exists.

        :param path: The path to check
        :return: True if the path exists, False for broken symbolic links otherwise
        """
        path = os.path.expanduser(path)
        return os.path.exists(path)

    def _create(self, path):
        """
        Creates a symbolic link to the specified path.

        :param path: The path to the symbolic link to create
        :return: True if the symbolic link has been created successfully, False otherwise
        """
        success = True
        parent = os.path.abspath(os.path.join(os.path.expanduser(path), os.pardir))
        if not self._exists(parent):
            try:
                os.makedirs(parent)
            except OSError:
                self._log.warning("Failed to create directory \"{}\"".format(parent))
                success = False
            else:
                self._log.lowinfo("Creating directory \"{}\"".format(parent))
        return success

    def _delete(self, source, path, relative, force):
        """
        Deletes the specified path.

        :param source: The link source
        :param path: The path to delete
        :param relative: Flag to indicate if the specified parameters are relative
        :param force: Flag to indicate if the deletion should be forced
        :return: True if the path has been deleted successfully, False otherwise
        """
        success = True
        source = os.path.join(self._context.snowblock_dir(), source)
        fullpath = os.path.expanduser(path)
        if relative:
            source = self._relative_path(source, fullpath)
        if (self._is_link(path) and self._link_destination(path) != source) or (self._exists(path) and not self._is_link(path)):
            removed = False
            try:
                if os.path.islink(fullpath):
                    os.unlink(fullpath)
                    removed = True
                elif force:
                    if os.path.isdir(fullpath):
                        shutil.rmtree(fullpath)
                        removed = True
                    else:
                        os.remove(fullpath)
                        removed = True
            except OSError:
                self._log.warning("Failed to remove {}".format(path))
                success = False
            else:
                if removed:
                    self._log.lowinfo("Removing {}".format(path))
        return success

    def _relative_path(self, source, destination):
        """
        Gets the relative path to get to the source path from the destination path.

        :param source: The source path
        :param destination: The destination path
        :return: The relative path
        """
        destination_dir = os.path.dirname(destination)
        return os.path.relpath(source, destination_dir)

    def _link(self, source, link_name, relative):
        """
        Links the specified link_name to the source.

        :param source: The source path to get linked
        :param link_name: The name of the link to link
        :return: True if the link has been linked successfully, False otherwise
        """
        success = False
        destination = os.path.expanduser(link_name)
        absolute_source = os.path.join(self._context.snowblock_dir(), source)
        if relative:
            source = self._relative_path(absolute_source, destination)
        else:
            source = absolute_source
        if not self._exists(link_name) and self._is_link(link_name) and self._link_destination(link_name) != source:
            self._log.warning("Invalid link {} -> {}".format(link_name, self._link_destination(link_name)))
        elif not self._exists(link_name) and self._exists(absolute_source):
            try:
                os.symlink(source, destination)
            except OSError:
                self._log.warning("Linking failed {} -> {}".format(link_name, source))
            else:
                self._log.lowinfo("Creating link {} -> {}".format(link_name, source))
                success = True
        elif self._exists(link_name) and not self._is_link(link_name):
            self._log.warning("{} already exists but is a regular file or directory".format(link_name))
        elif self._is_link(link_name) and self._link_destination(link_name) != source:
            self._log.warning("Incorrect link {} -> {}".format(link_name, self._link_destination(link_name)))
        elif not self._exists(absolute_source):
            if self._is_link(link_name):
                self._log.warning("Nonexistent target {} -> {}".format(link_name, source))
            else:
                self._log.warning("Nonexistent target for {} : {}".format(link_name, source))
        else:
            self._log.lowinfo("Link already exists {} -> {}".format(link_name, source))
            success = True
        return success
