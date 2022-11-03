package com.megaease.mesh.meshdemo;

import java.util.Arrays;
import java.net.InetAddress;
import com.fasterxml.jackson.databind.PropertyNamingStrategy;
import com.fasterxml.jackson.databind.annotation.JsonNaming;
import com.fasterxml.jackson.annotation.JsonInclude;
import lombok.Builder;
import lombok.Data;

@Data
@Builder
@JsonNaming(PropertyNamingStrategy.SnakeCaseStrategy.class)
public class RestaurantResponse {
    private String orderId;
    private String food;
    private String deliveryTime;
    private String[] serviceTracings;

    @JsonInclude(JsonInclude.Include.NON_EMPTY)
    private String coupon;

    public void updateServiceTracings(String[] serviceTracings) {
        String hostName = "";
        try {
            hostName = InetAddress.getLocalHost().getHostName();
        } catch (Exception e) {
            hostName = "restaurant-unknown";
        }

        this.serviceTracings = add2BeginningOfArray(serviceTracings, hostName);
    }

    public static <T> T[] add2BeginningOfArray(T[] elements, T element) {
        T[] newArray = Arrays.copyOf(elements, elements.length + 1);
        newArray[0] = element;
        System.arraycopy(elements, 0, newArray, 1, elements.length);

        return newArray;
    }
}
