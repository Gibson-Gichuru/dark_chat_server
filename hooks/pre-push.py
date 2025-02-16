#!/usr/bin/python3
import logging
import subprocess
import platform

logging.basicConfig(
    level=logging.DEBUG,
    format="[:] %(process)d - %(levelname)s - %(message)s"
)


def run_tests():

    logging.info("Running Application tests")

    subprocess.run("go test -v ./...".split(), check=True)


if __name__ == "__main__":

    run_tests()