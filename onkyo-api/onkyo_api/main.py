import logging
from dataclasses import dataclass
from typing import Optional

import eiscp
from fastapi import FastAPI, HTTPException

from onkyo_api.config import Profile, Settings, profile_friendly_name, profiles

MAX_RETRIES = 5
RETRY_COOLDOWN = 1
MAX_VOLUME = 40

logger = logging.getLogger("onkyo")


@dataclass
class OnkyoProxy:
    onkyo_host: str
    onkyo_port: int

    def command(self, cmd: str):
        if "system-power" not in cmd:
            self.set_power_on()

        for i in range(MAX_RETRIES):
            try:
                with eiscp.eISCP(self.onkyo_host, self.onkyo_port) as receiver:
                    receiver.CONNECT_TIMEOUT = 1
                    resp = receiver.command(cmd)
                    logger.info(f"Received response {resp} for command {cmd}")
                    return resp
            except ValueError as exc:
                if i == MAX_RETRIES - 1:
                    raise exc

    def raw(self, cmd: str):
        if "system-power" not in cmd:
            self.set_power_on()

        for i in range(MAX_RETRIES):
            try:
                with eiscp.eISCP(self.onkyo_host, self.onkyo_port) as receiver:
                    receiver.CONNECT_TIMEOUT = 1
                    resp = receiver.raw(cmd)
                    logger.info(f"Received response {resp} for command {cmd}")
                    return resp
            except ValueError as exc:
                if i == MAX_RETRIES - 1:
                    raise exc

    def is_powered(self) -> bool:
        resp = self.command("system-power=query")
        return resp[1] == "on"

    def set_power_on(self) -> bool:
        self.command("system-power=on")
        return True

    def set_power_off(self) -> bool:
        self.command("system-power=off")
        return False

    def switch_power(self) -> bool:
        if self.is_powered():
            return self.set_power_off()
        else:
            return self.set_power_on()

    def get_current_profile(self) -> Optional[Profile]:
        resp = self.command("input-selector=query")
        selector_name = ",".join(resp[1])

        if profile_name := profile_friendly_name.get(selector_name):
            return profiles[profile_name]
        else:
            return None

    def set_profile(self, profile_name: str) -> Optional[Profile]:
        if profile_name not in profiles:
            return None

        profile = profiles[profile_name]
        self.command(f"input-selector={profile.selector_name}")
        self.command(f"master-volume={profile.master_volume}")
        self.raw(f"SWL{profile.subwoofer_level}")
        return profile


app = FastAPI()
config = Settings()
onkyo = OnkyoProxy(config.onkyo_host, config.onkyo_port)


@app.get("/power")
def power_query():
    return {"is_powered": onkyo.is_powered()}


@app.put("/power/on")
def power_on():
    return {"is_powered": onkyo.set_power_on()}


@app.put("/power/off")
def power_off():
    return {"is_powered": onkyo.set_power_off()}


@app.put("/power/switch")
def power_switch():
    return {"is_powered": onkyo.switch_power()}


@app.get("/profile")
def profile_query():
    profile = onkyo.get_current_profile()
    if profile is None:
        return HTTPException(status_code=404, detail="Unknown profile")
    return profile


@app.put("/profile")
def select_profile(name: str):
    profile = onkyo.set_profile(name)
    if profile is None:
        return HTTPException(status_code=404, detail="Unknown profile")
    return profile


@app.get("/volume")
def volume_query():
    resp = onkyo.command("master-volume=query")
    return {"level": resp[1]}


@app.put("/volume")
def volume_set(value: int):
    if value > MAX_VOLUME:
        return None
    resp = onkyo.command(f"master-volume={value}")
    return {"level": resp[1]}


@app.put("/volume/up")
def volume_up():
    resp = onkyo.command("master-volume=level-up")
    return {"level": resp[1]}


@app.put("/volume/down")
def volume_down():
    resp = onkyo.command("master-volume=level-down")
    return {"level": resp[1]}


@app.get("/subwoofer")
def subwoofer_query():
    return {"level": int(onkyo.raw("SWLQSTN")[-3:])}


@app.put("/subwoofer")
def subwoofer_set(value: int):
    if not (-8 < value < 8):
        return HTTPException(status_code=404, detail="Subwoofer level must be between -8 and 8")

    if value < 0:
        value = f"-0{value}"
    else:
        value = f"+0{value}"

    onkyo.raw(f"SWL{value}")
    return {"level": int(value[-3:])}


@app.put("/subwoofer/up")
def subwoofer_up():
    return {"level": int(onkyo.raw("SWLUP")[-3:])}


@app.put("/subwoofer/down")
def subwoofer_down():
    return {"level": int(onkyo.raw("SWLDOWN")[-3:])}
