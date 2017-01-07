from distutils.core import setup

setup(
    name='snowsaw',
    version='0.0.0',
    packages=['', 'util', 'logging'],
    package_dir={'': 'snowsaw'},
    url='https://github.com/arcticicestudio/snowsaw',
    license='MIT',
    author='Arctic Ice Studio',
    author_email='development@arcticicestudio.com',
    description='A lightweight, plugin-driven and simple configurable dotfile bootstrapper.'
)
