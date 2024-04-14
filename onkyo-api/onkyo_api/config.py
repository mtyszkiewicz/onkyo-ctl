from pydantic import BaseModel
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    onkyo_host: str = "10.205.0.163"
    onkyo_port: int = 60128


class Profile(BaseModel):
    name: str
    selector: str
    master_volume: int
    subwoofer_level: int
    max_volume: int


profiles = {
    "tv": Profile(
        name="tv",
        selector="tv",
        master_volume=25,
        subwoofer_level=0,
        max_volume=40,
    ),
    "dj": Profile(
        name="dj",
        selector="dvd,bd,dvd",
        master_volume=30,
        subwoofer_level=-8,
        max_volume=40,
    ),
    "vinyl": Profile(
        name="vinyl",
        selector="phono",
        master_volume=25,
        subwoofer_level=0,
        max_volume=40,
    ),
    "spotify": Profile(
        name="spotify",
        selector="video2,cbl,sat",
        master_volume=34,
        subwoofer_level=-6,
        max_volume=40,
    )
}

profile_friendly_name = {
    "tv": "tv",
    "dvd,bd,dvd": "dj",
    "phono": "vinyl",
    "video2,cbl,sat": "spotify"
}
