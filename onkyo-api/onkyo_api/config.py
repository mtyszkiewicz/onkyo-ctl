from pydantic import BaseModel
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    onkyo_host: str = "10.205.0.163"
    onkyo_port: int = 60128


class Profile(BaseModel):
    friendly_name: str
    selector_name: str
    master_volume: int
    subwoofer_level: str


profiles = {
    "tv": Profile(
        friendly_name="tv",
        selector_name="tv",
        master_volume=25,
        subwoofer_level="+00",
        max_volume=40,
    ),
    "dj": Profile(
        friendly_name="dj",
        selector_name="dvd,bd,dvd",
        master_volume=30,
        subwoofer_level="-08",
        max_volume=40,
    ),
    "vinyl": Profile(
        friendly_name="vinyl",
        selector_name="phono",
        master_volume=25,
        subwoofer_level="+00",
        max_volume=40,
    ),
    "spotify": Profile(
        friendly_name="spotify",
        selector_name="video2,cbl,sat",
        master_volume=34,
        subwoofer_level="-06",
        max_volume=40,
    )
}

profile_friendly_name = {
    "tv": "tv",
    "dvd,bd,dvd": "dj",
    "phono": "vinyl",
    "video2,cbl,sat": "spotify"
}
