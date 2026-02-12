from spatial.v1 import types_pb2 as _types_pb2
from telemetry.v1 import telemetry_pb2 as _telemetry_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor
GAME_STATUS_PLAYING: GameStatus
GAME_STATUS_POST_MATCH: GameStatus
GAME_STATUS_POST_SUDDEN_DEATH: GameStatus
GAME_STATUS_PRE_MATCH: GameStatus
GAME_STATUS_PRE_SUDDEN_DEATH: GameStatus
GAME_STATUS_ROUND_OVER: GameStatus
GAME_STATUS_ROUND_START: GameStatus
GAME_STATUS_SCORE: GameStatus
GAME_STATUS_SUDDEN_DEATH: GameStatus
GAME_STATUS_UNSPECIFIED: GameStatus
MATCH_TYPE_ARENA: MatchType
MATCH_TYPE_COMBAT: MatchType
MATCH_TYPE_ECHO_PASS: MatchType
MATCH_TYPE_FFA: MatchType
MATCH_TYPE_PRIVATE: MatchType
MATCH_TYPE_SOCIAL_PRIVATE: MatchType
MATCH_TYPE_SOCIAL_PUBLIC: MatchType
MATCH_TYPE_TOURNAMENT: MatchType
MATCH_TYPE_UNSPECIFIED: MatchType
PAUSE_STATE_AUTOPAUSE_REPLAY: PauseState
PAUSE_STATE_NOT_PAUSED: PauseState
PAUSE_STATE_PAUSED: PauseState
PAUSE_STATE_UNPAUSING: PauseState
PAUSE_STATE_UNSPECIFIED: PauseState

class CaptureHeader(_message.Message):
    __slots__ = ["capture_id", "client_name", "created_at", "initial_roster", "map_name", "match_type", "metadata", "private_match", "session_id", "total_round_count", "tournament_match"]
    class MetadataEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    CAPTURE_ID_FIELD_NUMBER: _ClassVar[int]
    CLIENT_NAME_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    INITIAL_ROSTER_FIELD_NUMBER: _ClassVar[int]
    MAP_NAME_FIELD_NUMBER: _ClassVar[int]
    MATCH_TYPE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    PRIVATE_MATCH_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    TOTAL_ROUND_COUNT_FIELD_NUMBER: _ClassVar[int]
    TOURNAMENT_MATCH_FIELD_NUMBER: _ClassVar[int]
    capture_id: str
    client_name: str
    created_at: _timestamp_pb2.Timestamp
    initial_roster: _containers.RepeatedCompositeFieldContainer[PlayerInfo]
    map_name: str
    match_type: MatchType
    metadata: _containers.ScalarMap[str, str]
    private_match: bool
    session_id: str
    total_round_count: int
    tournament_match: bool
    def __init__(self, capture_id: _Optional[str] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., session_id: _Optional[str] = ..., map_name: _Optional[str] = ..., match_type: _Optional[_Union[MatchType, str]] = ..., client_name: _Optional[str] = ..., private_match: bool = ..., tournament_match: bool = ..., total_round_count: _Optional[int] = ..., initial_roster: _Optional[_Iterable[_Union[PlayerInfo, _Mapping]]] = ..., metadata: _Optional[_Mapping[str, str]] = ...) -> None: ...

class DiscState(_message.Message):
    __slots__ = ["bounce_count", "pose", "velocity"]
    BOUNCE_COUNT_FIELD_NUMBER: _ClassVar[int]
    POSE_FIELD_NUMBER: _ClassVar[int]
    VELOCITY_FIELD_NUMBER: _ClassVar[int]
    bounce_count: int
    pose: _types_pb2.Pose
    velocity: _types_pb2.Vec3
    def __init__(self, pose: _Optional[_Union[_types_pb2.Pose, _Mapping]] = ..., velocity: _Optional[_Union[_types_pb2.Vec3, _Mapping]] = ..., bounce_count: _Optional[int] = ...) -> None: ...

class EnvelopeV2(_message.Message):
    __slots__ = ["frame_v1", "frame_v2", "header_v1", "header_v2"]
    FRAME_V1_FIELD_NUMBER: _ClassVar[int]
    FRAME_V2_FIELD_NUMBER: _ClassVar[int]
    HEADER_V1_FIELD_NUMBER: _ClassVar[int]
    HEADER_V2_FIELD_NUMBER: _ClassVar[int]
    frame_v1: _telemetry_pb2.LobbySessionStateFrame
    frame_v2: Frame
    header_v1: _telemetry_pb2.TelemetryHeader
    header_v2: CaptureHeader
    def __init__(self, header_v2: _Optional[_Union[CaptureHeader, _Mapping]] = ..., frame_v2: _Optional[_Union[Frame, _Mapping]] = ..., header_v1: _Optional[_Union[_telemetry_pb2.TelemetryHeader, _Mapping]] = ..., frame_v1: _Optional[_Union[_telemetry_pb2.LobbySessionStateFrame, _Mapping]] = ...) -> None: ...

class Frame(_message.Message):
    __slots__ = ["blue_points", "disc", "disc_holder_slot", "events", "frame_index", "game_clock", "game_status", "orange_points", "pause_state", "player_bones", "players", "round_number", "timestamp_offset_ms", "vr_root"]
    BLUE_POINTS_FIELD_NUMBER: _ClassVar[int]
    DISC_FIELD_NUMBER: _ClassVar[int]
    DISC_HOLDER_SLOT_FIELD_NUMBER: _ClassVar[int]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    FRAME_INDEX_FIELD_NUMBER: _ClassVar[int]
    GAME_CLOCK_FIELD_NUMBER: _ClassVar[int]
    GAME_STATUS_FIELD_NUMBER: _ClassVar[int]
    ORANGE_POINTS_FIELD_NUMBER: _ClassVar[int]
    PAUSE_STATE_FIELD_NUMBER: _ClassVar[int]
    PLAYERS_FIELD_NUMBER: _ClassVar[int]
    PLAYER_BONES_FIELD_NUMBER: _ClassVar[int]
    ROUND_NUMBER_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_OFFSET_MS_FIELD_NUMBER: _ClassVar[int]
    VR_ROOT_FIELD_NUMBER: _ClassVar[int]
    blue_points: int
    disc: DiscState
    disc_holder_slot: int
    events: _containers.RepeatedCompositeFieldContainer[_telemetry_pb2.LobbySessionEvent]
    frame_index: int
    game_clock: float
    game_status: GameStatus
    orange_points: int
    pause_state: PauseState
    player_bones: _containers.RepeatedCompositeFieldContainer[PlayerBones]
    players: _containers.RepeatedCompositeFieldContainer[PlayerState]
    round_number: int
    timestamp_offset_ms: int
    vr_root: _types_pb2.Pose
    def __init__(self, frame_index: _Optional[int] = ..., timestamp_offset_ms: _Optional[int] = ..., game_status: _Optional[_Union[GameStatus, str]] = ..., game_clock: _Optional[float] = ..., disc: _Optional[_Union[DiscState, _Mapping]] = ..., players: _Optional[_Iterable[_Union[PlayerState, _Mapping]]] = ..., player_bones: _Optional[_Iterable[_Union[PlayerBones, _Mapping]]] = ..., events: _Optional[_Iterable[_Union[_telemetry_pb2.LobbySessionEvent, _Mapping]]] = ..., disc_holder_slot: _Optional[int] = ..., vr_root: _Optional[_Union[_types_pb2.Pose, _Mapping]] = ..., blue_points: _Optional[int] = ..., orange_points: _Optional[int] = ..., round_number: _Optional[int] = ..., pause_state: _Optional[_Union[PauseState, str]] = ...) -> None: ...

class PlayerBones(_message.Message):
    __slots__ = ["orientations", "slot", "transforms"]
    ORIENTATIONS_FIELD_NUMBER: _ClassVar[int]
    SLOT_FIELD_NUMBER: _ClassVar[int]
    TRANSFORMS_FIELD_NUMBER: _ClassVar[int]
    orientations: bytes
    slot: int
    transforms: bytes
    def __init__(self, slot: _Optional[int] = ..., transforms: _Optional[bytes] = ..., orientations: _Optional[bytes] = ...) -> None: ...

class PlayerInfo(_message.Message):
    __slots__ = ["account_number", "display_name", "role", "slot"]
    ACCOUNT_NUMBER_FIELD_NUMBER: _ClassVar[int]
    DISPLAY_NAME_FIELD_NUMBER: _ClassVar[int]
    ROLE_FIELD_NUMBER: _ClassVar[int]
    SLOT_FIELD_NUMBER: _ClassVar[int]
    account_number: int
    display_name: str
    role: _telemetry_pb2.Role
    slot: int
    def __init__(self, slot: _Optional[int] = ..., account_number: _Optional[int] = ..., display_name: _Optional[str] = ..., role: _Optional[_Union[_telemetry_pb2.Role, str]] = ...) -> None: ...

class PlayerState(_message.Message):
    __slots__ = ["body", "flags", "head", "left_hand", "ping", "right_hand", "slot", "velocity"]
    BODY_FIELD_NUMBER: _ClassVar[int]
    FLAGS_FIELD_NUMBER: _ClassVar[int]
    HEAD_FIELD_NUMBER: _ClassVar[int]
    LEFT_HAND_FIELD_NUMBER: _ClassVar[int]
    PING_FIELD_NUMBER: _ClassVar[int]
    RIGHT_HAND_FIELD_NUMBER: _ClassVar[int]
    SLOT_FIELD_NUMBER: _ClassVar[int]
    VELOCITY_FIELD_NUMBER: _ClassVar[int]
    body: _types_pb2.Pose
    flags: int
    head: _types_pb2.Pose
    left_hand: _types_pb2.Pose
    ping: int
    right_hand: _types_pb2.Pose
    slot: int
    velocity: _types_pb2.Vec3
    def __init__(self, slot: _Optional[int] = ..., head: _Optional[_Union[_types_pb2.Pose, _Mapping]] = ..., body: _Optional[_Union[_types_pb2.Pose, _Mapping]] = ..., left_hand: _Optional[_Union[_types_pb2.Pose, _Mapping]] = ..., right_hand: _Optional[_Union[_types_pb2.Pose, _Mapping]] = ..., velocity: _Optional[_Union[_types_pb2.Vec3, _Mapping]] = ..., flags: _Optional[int] = ..., ping: _Optional[int] = ...) -> None: ...

class GameStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []

class MatchType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []

class PauseState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
