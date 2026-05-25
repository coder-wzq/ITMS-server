-- ITMS: Create all 14 schemas for the microservice architecture
-- Postgres 16, DB: itms

-- CSCI-01: Authentication & Authorization
CREATE SCHEMA IF NOT EXISTS sch_auth;
-- CSCI-02: User Management
CREATE SCHEMA IF NOT EXISTS sch_user;
-- CSCI-03: System Configuration
CREATE SCHEMA IF NOT EXISTS sch_config;
-- CSCI-04: Director Control
CREATE SCHEMA IF NOT EXISTS sch_director;
-- CSCI-05: Task Management
CREATE SCHEMA IF NOT EXISTS sch_task;
-- CSCI-06: Task Planning
CREATE SCHEMA IF NOT EXISTS sch_planning;
-- CSCI-07: Mission Planning Tool
CREATE SCHEMA IF NOT EXISTS sch_mpt;
-- CSCI-09: Situation Display
CREATE SCHEMA IF NOT EXISTS sch_situation;
-- CSCI-10: Voice Communication
CREATE SCHEMA IF NOT EXISTS sch_voice;
-- CSCI-11: Data Dictionary
CREATE SCHEMA IF NOT EXISTS sch_dict;
-- CSCI-12: AI Agent
CREATE SCHEMA IF NOT EXISTS sch_agent;
-- CSCI-13: Analysis Report
CREATE SCHEMA IF NOT EXISTS sch_report;
-- CSCI-14: Training Record
CREATE SCHEMA IF NOT EXISTS sch_record;
-- Admin Management
CREATE SCHEMA IF NOT EXISTS sch_admin;
