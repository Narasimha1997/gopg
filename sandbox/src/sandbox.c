#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>
#include <fcntl.h>
#include <stdbool.h>


#define BUFFER_SIZE 4096
#define OUTPUT_BUFFER 1024

typedef unsigned char uchar;

void write_stdin_to_file(int * size) {

    uchar buffer[BUFFER_SIZE];
    int read_bytes = 0, itrs = 0;
    * size = 0;
    
    int fd = open("./binary", O_RDWR | O_CREAT, 0777);
    FILE * fp = fdopen(fd, "wb");

    if (fp == NULL) {
        fprintf(stdout, "Failed to open the file for writing\n");
        exit(-1);
    }
    
    while (true) {
        read_bytes = read(0, buffer, BUFFER_SIZE);
        if (read_bytes < 0) {
            fprintf(stdout, "Failed to read binary data, exiting");
            fclose(fp);
            exit(-1);
        }

        if (read_bytes == 0) {
            //EOF
            break;
        }

        //write data to the file
        *size = *size + read_bytes;
        fwrite(buffer, sizeof(uchar), read_bytes, fp);
    } 

    //wrote the file, close it.
    fclose(fp);
}

int main(int argc, char **argv) {
    int size = 0, fread_bytes = 0;

    char output_buffer[OUTPUT_BUFFER];
    write_stdin_to_file(&size);

    if (size == 0) {
        fprintf(stdout, "Empty binary file, discarding\n");
        exit(0);
    }

    FILE * process_fd = popen("./binary", "r");
    if (process_fd == NULL) {
        fprintf(stdout, "Failed to execute the binary\n");
        exit(-1);
    }

    //read the data as buffers and stream it to stdout
    while (true) {
        fread_bytes = fread(output_buffer, sizeof(uchar), sizeof(uchar) * OUTPUT_BUFFER, process_fd);

        if (fread_bytes == 0) {
            //EOF
            fclose(process_fd);
            exit(0);
        }

        if (fread_bytes < 0) {
            //Error 
            fprintf(stdout, "Failed to read the output");
            exit(-1);
        }

        output_buffer[fread_bytes] = '\0';

        fprintf(stdout, "%s", output_buffer);
    }

    return 0;
}