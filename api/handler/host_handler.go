package handler

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gringolito/dnsmasq-manager/api"
	"github.com/gringolito/dnsmasq-manager/api/dto"
	"github.com/gringolito/dnsmasq-manager/api/presenter"
	"github.com/gringolito/dnsmasq-manager/api/scope"
	"github.com/gringolito/dnsmasq-manager/api/validation"
	"github.com/gringolito/dnsmasq-manager/pkg/host"
	"github.com/gringolito/dnsmasq-manager/pkg/model"
	"golang.org/x/exp/slog"
)

// Error messages
const (
	StaticHostNotFoundMessage   = "No DHCP static host found for the given identifier."
	InvalidRequestMessage       = "The request is invalid."
	InvalidRequestBodyMessage   = "The request body was invalid."
	InvalidMacAddressMessage    = "The MAC address is invalid."
	DuplicatedMacAddressMessage = "A host with the same MAC address already exists."
	DuplicatedIPAddressMessage  = "The IP address is already in use."
)

// Details
const (
	NoMatchingIPAddress = "The DHCP server could not find a static host that matches the given IP address. " +
		"This could be because the host does not exist, or because the DHCP server is not configured to statically " +
		"assign an IP address to the given host. The IP address that was provided was: %s."
	NoMatchingMacAddress = "The DHCP server could not find a static host that matches the given MAC address. " +
		"This could be because the host does not exist, or because the DHCP server is not configured to statically " +
		"assign an IP address to the given host. The MAC address that was provided was: %s."
	MissingQueryParameter = "The request did not specify either the `mac` or `ip` query parameter. " +
		"Please specify one of these parameters in order to proceed."
	MalformedMacAddress = "The MAC address that was provided is not in the correct format. " +
		"Please check the format of the MAC address and try again. The MAC address that was provided was: %s."
	IPAddressAlreadyInUse = "The IP address that was provided is already in use by another host. " +
		"Please try again with a different IP address. The IP address that was provided was: %s."
	MacAddressAlreadyInUse = "The MAC address that was provided is already in use by another host: %s."
	HostCouldNotBeParsed   = "The request could not be processed because the host could not be parsed. Please check the request and try again."
)

func getHostFromBody(c *fiber.Ctx) *model.StaticDhcpHost {
	host := new(dto.StaticDhcpHost)
	if err := c.BodyParser(host); err != nil {
		slog.Debug("Failed to parse host from the body",
			slog.String("error", err.Error()),
		)
		presenter.UnprocessableEntityResponse(c, InvalidRequestBodyMessage, HostCouldNotBeParsed)
		return nil
	}

	if errors := validation.Validate(host); errors != nil {
		presenter.UnprocessableEntityResponse(c, InvalidRequestBodyMessage, errors)
		return nil
	}

	return host.ToModel()
}

func toStaticDhcpHostsDto(hosts *[]model.StaticDhcpHost) *[]dto.StaticDhcpHost {
	response := make([]dto.StaticDhcpHost, 0, len(*hosts))
	for _, h := range *hosts {
		response = append(response, *dto.NewStaticDhcpHost(&h))
	}

	return &response
}

func GetAllStaticHosts(service host.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		hosts, err := service.FetchAll()
		if err != nil {
			return presenter.InternalServerErrorResponse(c)
		}

		return c.Status(http.StatusOK).JSON(toStaticDhcpHostsDto(hosts))
	}
}

func GetStaticHost(service host.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		macAddress := c.Query("mac")
		if len(macAddress) > 0 {
			return getStaticHostByMac(service, c, macAddress)
		}

		ipAddress := c.Query("ip")
		if len(ipAddress) > 0 {
			return getStaticHostByIP(service, c, ipAddress)
		}

		return presenter.BadRequestResponse(c, InvalidRequestMessage, MissingQueryParameter)
	}
}

func getStaticHostByMac(service host.Service, c *fiber.Ctx, macAddress string) error {
	mac, err := net.ParseMAC(macAddress)
	if err != nil {
		slog.Debug("Could not parse MAC address",
			slog.String("macAddress", macAddress),
			slog.String("error", err.Error()),
		)
		return presenter.BadRequestResponse(c, InvalidMacAddressMessage, fmt.Sprintf(MalformedMacAddress, macAddress))
	}

	host, err := service.FetchByMac(mac)
	if err != nil {
		return presenter.InternalServerErrorResponse(c)
	}
	if host == nil {
		return presenter.NotFoundResponse(c, StaticHostNotFoundMessage, fmt.Sprintf(NoMatchingMacAddress, macAddress))
	}

	return c.Status(http.StatusOK).JSON(dto.NewStaticDhcpHost(host))
}

func getStaticHostByIP(service host.Service, c *fiber.Ctx, ipAddress string) error {
	host, err := service.FetchByIP(net.ParseIP(ipAddress))
	if err != nil {
		return presenter.InternalServerErrorResponse(c)
	}
	if host == nil {
		return presenter.NotFoundResponse(c, StaticHostNotFoundMessage, fmt.Sprintf(NoMatchingIPAddress, ipAddress))
	}

	return c.Status(http.StatusOK).JSON(dto.NewStaticDhcpHost(host))
}

func AddStaticHost(service host.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := getHostFromBody(c)
		if h == nil {
			// The errors was already handled by the getHostFromBody()
			return nil
		}

		if err := service.Insert(h); err != nil {
			if e, ok := err.(host.DuplicatedEntryError); ok {
				slog.Debug("Could not add a new static host because a conflict was detected",
					slog.Any("host", h),
					slog.String("error", err.Error()),
				)
				if e.Field == "IP" {
					return presenter.ConflictResponse(c, DuplicatedIPAddressMessage, fmt.Sprintf(IPAddressAlreadyInUse, h.IPAddress.String()))
				} else {
					return presenter.ConflictResponse(c, DuplicatedMacAddressMessage, fmt.Sprintf(MacAddressAlreadyInUse, h.MacAddress.String()))
				}
			} else {
				return presenter.InternalServerErrorResponse(c)
			}
		}

		return c.Status(http.StatusCreated).JSON(dto.NewStaticDhcpHost(h))
	}
}

func UpdateStaticHost(service host.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		host := getHostFromBody(c)
		if host == nil {
			// The errors was already handled by the getHostFromBody()
			return nil
		}

		if err := service.Update(host); err != nil {
			return presenter.InternalServerErrorResponse(c)
		}

		return c.Status(http.StatusCreated).JSON(dto.NewStaticDhcpHost(host))
	}
}

func RemoveStaticHost(service host.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		macAddress := c.Query("mac")
		if len(macAddress) > 0 {
			return removeStaticHostByMac(service, c, macAddress)
		}

		ipAddress := c.Query("ip")
		if len(ipAddress) > 0 {
			return removeStaticHostByIP(service, c, ipAddress)
		}

		return presenter.BadRequestResponse(c, InvalidRequestMessage, MissingQueryParameter)
	}
}

func removeStaticHostByMac(service host.Service, c *fiber.Ctx, macAddress string) error {
	mac, err := net.ParseMAC(macAddress)
	if err != nil {
		slog.Debug("Could not parse MAC address",
			slog.String("macAddress", macAddress),
			slog.String("error", err.Error()),
		)
		return presenter.BadRequestResponse(c, InvalidMacAddressMessage, fmt.Sprintf(MalformedMacAddress, macAddress))
	}

	host, err := service.RemoveByMac(mac)
	if err != nil {
		return presenter.InternalServerErrorResponse(c)
	}
	if host == nil {
		return c.SendStatus(http.StatusNoContent)
	}

	return c.Status(http.StatusOK).JSON(dto.NewStaticDhcpHost(host))
}

func removeStaticHostByIP(service host.Service, c *fiber.Ctx, ipAddress string) error {
	host, err := service.RemoveByIP(net.ParseIP(ipAddress))
	if err != nil {
		return presenter.InternalServerErrorResponse(c)
	}
	if host == nil {
		return c.SendStatus(http.StatusNoContent)
	}

	return c.Status(http.StatusOK).JSON(dto.NewStaticDhcpHost(host))
}

func RouteStaticHosts(router api.Router, service host.Service) {
	router.AddApiV1Route("/static", func(r fiber.Router) {
		r.Get("/hosts", router.AuthenticationHandler(scope.DhcpCanRead...), GetAllStaticHosts(service)).Name("get_all")
		r.Get("/host", router.AuthenticationHandler(scope.DhcpCanRead...), GetStaticHost(service)).Name("get")
		r.Post("/host", router.AuthenticationHandler(scope.DhcpCanAdd...), AddStaticHost(service)).Name("add")
		r.Put("/host", router.AuthenticationHandler(scope.DhcpCanChange...), UpdateStaticHost(service)).Name("update")
		r.Delete("/host", router.AuthenticationHandler(scope.DhcpCanChange...), RemoveStaticHost(service)).Name("remove")
	}, "static.hosts.")
}
