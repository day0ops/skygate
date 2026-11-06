export interface Meta {
  version: string;
  color: string;
  buildVersion: string;
}

export interface Gate {
  code: string;
}

export interface Airport {
  code: string;
  name: string;
  city: string;
  x: number;
  y: number;
  gates: Gate[];
}

export type FlightStatus =
  | "scheduled"
  | "boarding"
  | "departed"
  | "airborne"
  | "landed"
  | "delayed"
  | "cancelled";

export interface Position {
  x: number;
  y: number;
  heading: number;
  progress: number;
}

export interface Flight {
  id: string;
  flightNumber: string;
  originCode: string;
  destinationCode: string;
  aircraftId: string;
  status: FlightStatus;
  scheduledDeparture: string;
  scheduledArrival: string;
  gate?: string;
  position?: Position;
  delayMinutes?: number;
}

export interface FlightUpdateMessage {
  type: "flight_update";
  flight: Flight;
}

export interface LogEntry {
  id: number;
  text: string;
}
