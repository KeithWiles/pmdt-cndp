################################################################################
# Automatically-generated file. Do not edit!
################################################################################

# Add inputs and outputs from these tool invocations to the build variables
CPP_SRCS += \
common.cpp \
pcm-info.cpp \
pinfo.c \
system-info.c \
main.cpp


OBJS += \
./pcm-info.o \
./pinfo.o \
./system-info.o \
./main.o

CPP_DEPS += \
./common.d \
./pcm-info.d \
./pinfo.d \
./system-info.d \
./main.d


# Each subdirectory must supply rules for building sources it contributes
%.o: ../%.cpp
	@echo 'Building file: $<'
	@echo 'Invoking: GCC C++ Compiler'
	g++ -O0 -g3 -Wall -c -fmessage-length=0 -Wno-unknown-pragmas -std=c++0x -MMD -MP -MF"$(@:%.o=%.d)" -MT"$(@:%.o=%.d)" -o "$@" "$<"
	@echo 'Finished building: $<'
	@echo ' '
