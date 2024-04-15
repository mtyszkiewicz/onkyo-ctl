import logging
from dataclasses import dataclass
from time import sleep

import eiscp
from fastapi import FastAPI, HTTPException
from onkyo_api.config import (DeviceInfo, Settings, profile_friendly_name,
                              profiles)

MAX_RETRIES = 5
RETRY_COOLDOWN = 1
MAX_VOLUME = 50

logger = logging.getLogger("onkyo")


def level_format(level: int) -> str:
    sign_prefix = "-" if level < 0 else "+"
    return sign_prefix + str(abs(level)).zfill(2)


def level_parse(resp: str) -> int:
    return int(resp[-3:])


@dataclass
class OnkyoProxy:
    onkyo_host: str
    onkyo_port: int

    def command(self, cmd: str):
        for i in range(MAX_RETRIES):
            try:
                with eiscp.eISCP(self.onkyo_host, self.onkyo_port) as receiver:
                    receiver.CONNECT_TIMEOUT = 1
                    if "=" in cmd:
                        resp = receiver.command(cmd)
                    else:
                        resp = receiver.raw(cmd)
                    logger.info(f"Received response {resp} for command {cmd}")
                    return resp
            except ValueError as exc:
                sleep(0.5)
                if i == MAX_RETRIES - 1:
                    raise exc
            except AssertionError:
                raise HTTPException(429)

    def get_device_info(self) -> DeviceInfo:
        profile_name = self.get_profile_name()
        if profile_name in profiles:
            max_volume = profiles[profile_name].max_volume
        else:
            max_volume = MAX_VOLUME

        return DeviceInfo(
            profile=profile_name,
            volume_level=self.get_volume(),
            subwoofer_level=self.get_subwoofer_level(),
            max_volume=max_volume,
        )

    def is_powered(self) -> bool:
        resp = self.command("system-power=query")
        return resp[1] == "on"

    def power_on(self) -> bool:
        self.command("system-power=on")
        return True

    def power_off(self) -> bool:
        self.command("system-power=off")
        return False

    def switch_power(self) -> bool:
        if self.is_powered():
            return self.power_off()
        else:
            return self.power_on()

    def get_volume(self) -> int:
        resp = onkyo.command("master-volume=query")
        return resp[1]

    def set_volume(self, value: int) -> int:
        resp = onkyo.command(f"master-volume={value}")
        return resp[1]

    def volume_up(self) -> int:
        resp = self.command("master-volume=level-up")
        return resp[1]

    def volume_down(self) -> int:
        resp = self.command("master-volume=level-down")
        return resp[1]

    def get_subwoofer_level(self) -> int:
        return level_parse(self.command("SWLQSTN"))

    def set_subwoofer_level(self, level) -> int:
        return level_parse(self.command(f"SWL{level_format(level)}"))

    def subwoofer_level_up(self) -> int:
        return level_parse(self.command("SWLUP"))

    def subwoofer_level_down(self) -> int:
        return level_parse(self.command("SWLDOWN"))

    def get_input_selector(self) -> str:
        resp = self.command("input-selector=query")[1]
        return ",".join(resp) if isinstance(resp, tuple) else resp

    def set_input_selector(self, selector: str) -> str:
        resp = self.command(f"input-selector={selector}")[1]
        return ",".join(resp) if isinstance(resp, tuple) else resp

    def get_profile_name(self) -> str:
        selector = self.get_input_selector()
        return profile_friendly_name.get(selector, "unknown")

    def set_profile(self, profile_name: str) -> DeviceInfo:
        if profile_name not in profiles:
            return None

        profile = profiles[profile_name]
        self.set_input_selector(profile.selector)
        self.set_volume(profile.volume_level)
        self.set_subwoofer_level(profile.subwoofer_level)

        return DeviceInfo(
            profile=profile.name,
            volume_level=profile.volume_level,
            subwoofer_level=profile.subwoofer_level,
            max_volume=profile.max_volume,
        )


app = FastAPI()
config = Settings()
onkyo = OnkyoProxy(config.onkyo_host, config.onkyo_port)


@app.get("/power")
def power_query():
    return {"is_powered": onkyo.is_powered()}


@app.put("/power/on")
def power_on():
    return {"is_powered": onkyo.power_on()}


@app.put("/power/off")
def power_off():
    return {"is_powered": onkyo.power_off()}


@app.put("/power/switch")
def power_switch():
    return {"is_powered": onkyo.switch_power()}


@app.get("/device")
def profile_query():
    return onkyo.get_device_info()


@app.put("/profile")
def select_profile(name: str):
    # return onkyo.set_profile(name)
    onkyo.set_profile(name)
    return onkyo.get_device_info()


@app.get("/volume")
def volume_query():
    return {"level": onkyo.get_volume()}


@app.put("/volume")
def volume_set(level: int):
    if level > MAX_VOLUME:
        level = MAX_VOLUME
    onkyo.set_volume(level)

    return onkyo.get_device_info()


@app.put("/volume/up")
def volume_up():
    onkyo.volume_up()
    return onkyo.get_device_info()


@app.put("/volume/down")
def volume_down():
    onkyo.volume_down()
    return onkyo.get_device_info()


@app.get("/subwoofer")
def subwoofer_query():
    return {"level": onkyo.get_subwoofer_level()}


@app.put("/subwoofer")
def subwoofer_set(level: int):
    onkyo.set_subwoofer_level(level)
    return onkyo.get_device_info()


@app.put("/subwoofer/up")
def subwoofer_up():
    onkyo.subwoofer_level_up()
    return onkyo.get_device_info()


@app.put("/subwoofer/down")
def subwoofer_down():
    onkyo.subwoofer_level_down()
    return onkyo.get_device_info()
