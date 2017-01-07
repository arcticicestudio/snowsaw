"""
The snowsaw CLI.

This is the main entry point of the public API.
"""
from argparse import ArgumentParser
import glob
import os

from .config import ConfigReader, ReadingError
from .dispatcher import Dispatcher, DispatchError
from .logging import Level
from .logging import Logger
from .util import module


def add_options(parser):
    """
    Adds all options to the specified parser.

    :param parser: The parser to add all options to
    :return: None
    """
    parser.add_argument('-Q', '--super-quiet', dest='super_quiet', action='store_true', help='suppress almost all output')
    parser.add_argument('-q', '--quiet', dest='quiet', action='store_true', help='suppress most output')
    parser.add_argument('-v', '--verbose', dest='verbose', action='store_true', help='enable verbose output')
    parser.add_argument('-s', '--snowblocks-directory', nargs=1, dest='snowblocks_directory',
                        help='base snowblock directory to run all tasks of', metavar='SNOWBLOCKSDIR', required=True)
    parser.add_argument('-c', '--config-file', nargs=1, dest='config_file', help='run tasks for the specified snowblock', metavar='CONFIGFILE')
    parser.add_argument('-p', '--plugin', action='append', dest='plugins', default=[], help='load PLUGIN as a plugin', metavar='PLUGIN')
    parser.add_argument('--disable-core-plugins', dest='disable_core_plugins', action='store_true', help='disable all core plugins')
    parser.add_argument('--plugin-dir', action='append', dest='plugin_dirs', default=[], metavar='PLUGIN_DIR', help='load all plugins in PLUGIN_DIR')


def read_config(config_file):
    """
    Reads the specified configuration file.

    :param config_file: The configuration file to read
    :return: The read configuration data
    """
    reader = ConfigReader(config_file)
    return reader.get_config()


def main():
    """
    Processes all parsed options and hands it over to the dispatcher for each snowblock.

    :return: True if all tasks have been executed successfully, False otherwise
    """
    log = Logger()
    try:
        parser = ArgumentParser()
        snowblock_config_filename = "snowblock.json"
        add_options(parser)
        options = parser.parse_args()

        if options.super_quiet:
            log.set_level(Level.WARNING)
        if options.quiet:
            log.set_level(Level.INFO)
        if options.verbose:
            log.set_level(Level.DEBUG)

        plugin_directories = list(options.plugin_dirs)
        if not options.disable_core_plugins:
            plugin_directories.append(os.path.join(os.path.dirname(__file__), "plugins"))
        plugin_paths = []
        for directory in plugin_directories:
            for plugin_path in glob.glob(os.path.join(directory, "*.py")):
                plugin_paths.append(plugin_path)
        for plugin_path in options.plugins:
            plugin_paths.append(plugin_path)
        for plugin_path in plugin_paths:
            abspath = os.path.abspath(plugin_path)
            module.load(abspath)

        if options.config_file:
            snowblocks = [os.path.basename(os.path.dirname(options.config_file[0]))]
        else:
            snowblocks = [snowblock for snowblock in os.listdir(options.snowblocks_directory[0])
                          if os.path.isdir(os.path.join(options.snowblocks_directory[0], snowblock))]

        for snowblock in snowblocks:
            if os.path.isfile(os.path.join(snowblock, snowblock_config_filename)):
                log.info("❄ {}".format(snowblock))
                tasks = read_config(os.path.join(snowblock, snowblock_config_filename))

                if not isinstance(tasks, list):
                    raise ReadingError("Configuration file must be a list of tasks")

                dispatcher = Dispatcher(snowblock)
                success = dispatcher.dispatch(tasks)
                if success:
                    log.info("==> All tasks executed successfully\n")
                else:
                    raise DispatchError("\n==> Some tasks were not executed successfully")
            else:
                log.lowinfo("Skipped snowblock \"{}\": No configuration file found".format(snowblock))
    except (ReadingError, DispatchError) as e:
        log.error("{}".format(e))
        exit(1)
    except KeyboardInterrupt:
        log.error("\n==> Operation aborted")
        exit(1)
