from onkyo_api.utils import to_camel
from pydantic import BaseModel
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    onkyo_host: str = "10.205.0.163"
    onkyo_port: int = 60128


class Profile(BaseModel):
    name: str
    selector: str
    volume_level: int
    subwoofer_level: int
    max_volume: int = 50

class DeviceInfo(BaseModel):
    profile: str
    volume_level: int
    subwoofer_level: int
    max_volume: int

    class Config:
        populate_by_name = True
        alias_generator = to_camel


profiles = {
    "tv": Profile(
        name="tv",
        selector="tv",
        volume_level=20,
        subwoofer_level=0,
        max_volume=28,
    ),
    "dj": Profile(
        name="dj",
        selector="dvd,bd,dvd",
        volume_level=27,
        subwoofer_level=-8,
        max_volume=35,
    ),
    "vinyl": Profile(
        name="vinyl",
        selector="phono",
        volume_level=20,
        subwoofer_level=0,
        max_volume=30,
    ),
    "spotify": Profile(
        name="spotify",
        selector="video2,cbl,sat",
        volume_level=38,
        subwoofer_level=-6,
        max_volume=50,
    )
}

profile_friendly_name = {
    "tv": "tv",
    "dvd,bd,dvd": "dj",
    "phono": "vinyl",
    "video2,cbl,sat": "spotify"
}
